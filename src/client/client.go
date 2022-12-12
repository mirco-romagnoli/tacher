package client

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"tacher/src/model"
	"tacher/src/utils"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

const SPRING_URL = "https://start.spring.io/"

// generates the project package from the given data
func Generate(data *model.AppData) error {
	req, err := http.NewRequest("GET", SPRING_URL+"starter.zip", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("type", data.SpringBuildTool)
	q.Add("language", data.Language)
	q.Add("bootVersion", data.SpringBootVersion)
	q.Add("baseDir", data.Artifact)
	q.Add("groupId", data.Group)
	q.Add("artifactId", data.Artifact)
	q.Add("name", data.Name)
	q.Add("description", data.Description)
	q.Add("packageName", data.Pkg)
	q.Add("packaging", data.Packaging)
	q.Add("javaVersion", data.JavaVersion)
	q.Add("dependencies", strings.Join(utils.Map(data.Dependencies, func(v model.ValueWithDesc) string { return v.ID }), ","))
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer utils.CheckClose(resp.Body)

	if resp.StatusCode != 200 {
		// if the response contains a body then try to parse it to return a better feedback
		if resp.Body != nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("unexpected response code [%d]. Can't read error message [%w]", resp.StatusCode, err)
			}
			errorMessage, err := getErrorMessageFromResponse(body)
			if err != nil {
				return fmt.Errorf("unexpected response code [%d]. Can't parse error message [%w]", resp.StatusCode, err)
			}
			return fmt.Errorf("unexpected response code [%d]. Message: [%s]", resp.StatusCode, errorMessage)
		}
		// no body in the message. return generic error
		return fmt.Errorf("unexpected response code [%d].", resp.StatusCode)
	}

	// read the response
	archive, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := unzip(archive, data.Path); err != nil {
		return err
	}

	return nil
}

// gets the options from Spring initializer and puts them in the app's state
func GetOptions(state *model.AppState) error {
	// get data from Spring's website
	resp, err := http.Get(SPRING_URL + "metadata/client")
	if err != nil {
		return err
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	obj, err := oj.Parse(response)
	if err != nil {
		return err
	}

	// fill text defaults
	textDefaults := map[string]*string{
		"$.groupId.default":     &state.DefaultGroupId,
		"$.artifactId.default":  &state.DefaultArtifactId,
		"$.version.default":     &state.DefaultVersion,
		"$.name.default":        &state.DefaultName,
		"$.description.default": &state.DefaultDescription,
		"$.packageName.default": &state.DefaultPackageName,
	}
	for k, v := range textDefaults {
		if def, err := extract[string](obj, k, func(i []interface{}) interface{} { return i[0] }); err == nil {
			*v = def
		}
	}

	// fill Spring build tools
	state.SpringBuildTools, err = extract[[]model.ValueWithDesc](obj, "$.type.values[*]", nil)
	if err != nil {
		return err
	}
	if def, err := extract[string](obj, "$.type.default", func(i []interface{}) interface{} { return i[0] }); err == nil {
		if idx, found := utils.Find(state.SpringBuildTools, func(st model.ValueWithDesc) bool { return st.ID == def }); found {
			state.DefaultSpringBuildTool = idx
		} else {
			state.DefaultSpringBuildTool = 0
		}
	} else {
		return err
	}

	// fill packaging
	state.Packaging, err = extract[[]model.Value](obj, "$.packaging.values[*]", nil)
	if err != nil {
		return err
	}
	if def, err := extract[string](obj, "$.packaging.default", func(i []interface{}) interface{} { return i[0] }); err == nil {
		if idx, found := utils.Find(state.Packaging, func(p model.Value) bool { return p.ID == def }); found {
			state.DefaultPackaging = idx
		} else {
			state.DefaultPackaging = 0
		}
	} else {
		return err
	}

	// fill Java versions
	state.JavaVersions, err = extract[[]model.Value](obj, "$.javaVersion.values[*]", nil)
	if err != nil {
		return err
	}
	if def, err := extract[string](obj, "$.javaVersion.default", func(i []interface{}) interface{} { return i[0] }); err == nil {
		if idx, found := utils.Find(state.JavaVersions, func(v model.Value) bool { return v.ID == def }); found {
			state.DefaultJavaVersion = idx
		} else {
			state.DefaultJavaVersion = 0
		}
	} else {
		return err
	}

	// fill languages
	state.Languages, err = extract[[]model.Value](obj, "$.language.values[*]", nil)
	if err != nil {
		return err
	}
	if def, err := extract[string](obj, "$.language.default", func(i []interface{}) interface{} { return i[0] }); err == nil {
		if idx, found := utils.Find(state.Languages, func(l model.Value) bool { return l.ID == def }); found {
			state.DefaultLanguage = idx
		} else {
			state.DefaultLanguage = 0
		}
	} else {
		return err
	}

	// fill Spring Boot versions
	state.SpringVersions, err = extract[[]model.Value](obj, "$.bootVersion.values[*]", nil)
	if err != nil {
		return err
	}
	if def, err := extract[string](obj, "$.bootVersion.default", func(i []interface{}) interface{} { return i[0] }); err == nil {
		if idx, found := utils.Find(state.SpringVersions, func(v model.Value) bool { return v.ID == def }); found {
			state.DefaultSpringVersion = idx
		} else {
			state.DefaultSpringVersion = 0
		}
	} else {
		return err
	}

	// fill dependencies
	state.Dependency, err = extractDependencies(obj)
	if err != nil {
		return err
	}

	return nil
}

// unzip the archive in the given directory
func unzip(archive []byte, dest string) error {
	arch, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return err
	}

	for _, f := range arch.File {
		filePath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("can't create directory for %s. %w", filePath, err)
		}

		dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		archiveFile, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dst, archiveFile); err != nil {
			return err
		}

		dst.Close()
		archiveFile.Close()

	}
	return nil
}

// extract the dependencies from Spring intializer's response, returns a map where
// each key is the category and the values are the dependencies in that category
func extractDependencies(obj interface{}) (map[string][]model.ValueWithDesc, error) {
	path, err := jp.ParseString("$.dependencies.values[*]")
	if err != nil {
		return nil, err
	}
	values := path.Get(obj)
	ret := make(map[string][]model.ValueWithDesc)
	for _, value := range values {
		name := value.(map[string]interface{})["name"].(string)
		ret[name], err = extract[[]model.ValueWithDesc](value, "$.values[*]", nil)
		sort.Sort(model.ValueWithDescByName(ret[name]))
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// map the result of the json path on the input interface and maps it on the given type
func extract[T interface{}](obj interface{}, jsonPath string, mergeFunction func([]interface{}) interface{}) (T, error) {
	if mergeFunction == nil {
		// if no merge function was passed then return the identity function
		mergeFunction = func(i []interface{}) interface{} { return i }
	}

	path, err := jp.ParseString(jsonPath)
	if err != nil {
		var zero T
		return zero, err
	}
	results := mergeFunction(path.Get(obj))
	json := oj.JSON(results)
	var ret T
	oj.Unmarshal([]byte(json), &ret)
	return ret, nil
}

// extract the error message from Spring intializer's response
func getErrorMessageFromResponse(response []byte) (string, error) {
	parsed, err := oj.Parse(response)
	if err != nil {
		return "", err
	}
	path, err := jp.ParseString("$.message")
	if err != nil {
		return "", err
	}
	return path.Get(parsed)[0].(string), nil
}
