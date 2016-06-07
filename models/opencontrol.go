package models

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"github.com/opencontrol/compliance-masonry/models/components"
	"github.com/opencontrol/compliance-masonry/models/components/versions"
	"github.com/opencontrol/compliance-masonry/models/components/versions/base"
)

var (
	// ErrReadDir is raised when a directory can not be read
	ErrReadDir = errors.New("Unable to read the directory")
	// ErrReadFile is raised when a file can not be read
	ErrReadFile = errors.New("Unable to read the file")
	// ErrCertificationSchema is raised a certification cannot be parsed
	ErrCertificationSchema = errors.New("Unable to parse certification")
	// ErrControlSchema is raised a control cannot be parsed
	ErrControlSchema = errors.New("Unable to parse control")
	// ErrStandardSchema is raised a standard cannot be parsed
	ErrStandardSchema = errors.New("Unable to parse standard")
	// ErrComponentFileDNE is raised when a component file does not exists
	ErrComponentFileDNE = errors.New("Component files does not exist")
)

// OpenControl struct combines components, standards, and a certification data
// For more information on the opencontrol schema visit: https://github.com/opencontrol/schemas
type OpenControl struct {
	Components     *components.Components
	Standards      *Standards
	Justifications *Justifications
	Certification  *Certification
}

// getKey extracts a component key from the filepath
func getKey(filePath string) string {
	_, key := filepath.Split(filePath)
	return key
}

// NewOpenControl initializes an empty OpenControl struct
func NewOpenControl() *OpenControl {
	return &OpenControl{
		Justifications: NewJustifications(),
		Components:     components.NewComponents(),
		Standards:      NewStandards(),
	}
}

// LoadData creates a new instance of OpenControl struct and loads
// the components, standards, and certification data.
func LoadData(openControlDir string, certificationPath string) *OpenControl {
	var wg sync.WaitGroup
	openControl := NewOpenControl()
	wg.Add(3)
	go func() {
		defer wg.Done()
		openControl.LoadComponents(filepath.Join(openControlDir, "components"))
	}()
	go func() {
		defer wg.Done()
		openControl.LoadStandards(filepath.Join(openControlDir, "standards"))
	}()
	go func() {
		defer wg.Done()
		openControl.LoadCertification(certificationPath)
	}()
	wg.Wait()
	return openControl
}

// LoadComponents loads multiple components by searching for components in a
// given directory
func (openControl *OpenControl) LoadComponents(directory string) error {
	var wg sync.WaitGroup
	componentsDir, err := ioutil.ReadDir(directory)
	if err != nil {
		return ErrReadDir
	}
	for _, componentDir := range componentsDir {
		wg.Add(1)
		go func(componentDir os.FileInfo) {
			if componentDir.IsDir() {
				componentDir := filepath.Join(directory, componentDir.Name())
				openControl.LoadComponent(componentDir)
			}
			wg.Done()
		}(componentDir)
	}
	wg.Wait()
	return nil
}

// LoadStandards loads multiple standards by searching for components in a
// given directory
func (openControl *OpenControl) LoadStandards(standardsDir string) error {
	var wg sync.WaitGroup

	standardsFiles, err := ioutil.ReadDir(standardsDir)
	if err != nil {
		return ErrReadDir
	}
	for _, standardFile := range standardsFiles {
		wg.Add(1)
		go func(standardFile os.FileInfo) {
			openControl.LoadStandard(filepath.Join(standardsDir, standardFile.Name()))
			wg.Done()
		}(standardFile)
	}
	wg.Wait()
	return nil
}

// LoadComponent imports components into a Component struct and adds it to the
// Components map.
func (openControl *OpenControl) LoadComponent(componentDir string) error {
	fileName := filepath.Join(componentDir, "component.yaml")
	_, err := os.Stat(fileName)
	if err != nil {
		return ErrComponentFileDNE
	}
	var component base.Component
	componentData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ErrReadFile
	}
	component, err = versions.ParseComponent(componentData)
	if err != nil {
		return err
	}

	/*
	err = yaml.Unmarshal(componentData, &component)
	// If we have a user friendly error via componentLoadError return it.
	if err != nil {
		switch errValue := err.(type) {
		// If we a user friendly error, let's return it now.
		case componentLoadError:
			return errValue
		}
	}
	// If we don't have a user friendly error yet...
	// Check the component version to give a better error before the generic "ErrControlSchema"
	if versionErr := component.VerifySchemaCompatibility(fileName); versionErr != nil {
		return versionErr
	}

	// If no specific errors were found, but a general error was found, return that.
	if err != nil {
		return ErrControlSchema
	}
	*/

	if component.GetKey() == "" {
		component.SetKey(getKey(componentDir))
	}
	if openControl.Components.CompareAndAdd(component) {
		openControl.Justifications.LoadMappings(component)
	}
	return nil
}