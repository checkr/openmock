package admin

import (
	"encoding/json"

	"github.com/checkr/openmock"
	model "github.com/checkr/openmock/swagger_gen/models"
	yamlToJSON "github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

func openmockToSwaggerMockArray(om []*openmock.Mock) (swagger model.Mocks) {
	mocks := make([]*model.Mock, 0, len(om))
	for _, mock := range om {
		swaggerMock := openmockToSwaggerMock(mock)
		mocks = append(mocks, &swaggerMock)
	}
	return mocks
}

func swaggerToOpenmockMockArray(swagger model.Mocks) (om []*openmock.Mock) {
	mocks := make([]*openmock.Mock, 0, len(swagger))
	for _, mock := range swagger {
		omMock := swaggerToOpenmockMock(*mock)
		mocks = append(mocks, &omMock)
	}
	return mocks
}

func openmockToSwaggerMock(om *openmock.Mock) (swagger model.Mock) {
	yamlBytes, err := yaml.Marshal(om)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"mock": om,
		}).Errorf("Failed to marshal open mock mock to yaml %v", err)
		return swagger
	}
	yamlText := string(yamlBytes)

	jsonBytes, err := yamlToJSON.YAMLToJSON([]byte(yamlText))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"yamlText": yamlText,
		}).Errorf("Failed to convert YAML to JSON %v", err)
		return swagger
	}
	jsonText := string(jsonBytes)

	err = json.Unmarshal([]byte(jsonText), &swagger)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"jsonText": jsonText,
			"yamlText": yamlText,
		}).Errorf("Failed to unmarshal JSON to swagger mock %v", err)
	}

	return swagger
}

// use producer?
func swaggerToOpenmockMock(swagger model.Mock) (om openmock.Mock) {
	jsonBytes, err := json.Marshal(swagger)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"mock": swagger,
		}).Errorf("Failed to marshal swagger mock to JSON %v %v", err, swagger)
		return om
	}
	jsonText := string(jsonBytes)

	yamlBytes, err := yamlToJSON.JSONToYAML([]byte(jsonText))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"jsonText": jsonText,
		}).Errorf("Failed to convert JSON to YAML %v", err)
		return om
	}
	yamlText := yamlBytes

	err = yaml.UnmarshalStrict(yamlText, &om)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"yamlText": yamlText,
		}).Errorf("Failed to unmarshal openmock mock from YAML %v", err)
	}
	return om
}
