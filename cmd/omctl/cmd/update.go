package cmd

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/checkr/openmock"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push local openmock model to remote instance",
	Long:  "push local openmock model to remote instance",
	Args:  cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// check if remote instance reachable
		response, httpErr := http.Get(openMockURL + "/api/v1/templates")
		if httpErr != nil || response == nil {
			return fmt.Errorf("Remote openmock instance not reachable %s: [%s]", openMockURL, httpErr)
		}
		if response.StatusCode != 200 {
			return fmt.Errorf("Remote openmock instance not reachable %s: [%d]", openMockURL, response.StatusCode)
		}
		response.Body.Close()
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// load local templates
		localOpenMock := openmock.OpenMock{}
		localOpenMock.ParseEnv()
		localOpenMock.TemplatesDir = localDirectory

		localOpenMock.SetupLogrus()
		localOpenMock.SetupRepo()

		err := localOpenMock.Load()
		if err != nil {
			logrus.Errorf("%s: %s", "failed to load yaml templates for mocks", err)
			return err
		}

		httpPath := "/api/v1/templates"
		if setKey != "" {
			httpPath = "/api/v1/template_sets/" + setKey
		}

		// post the loaded templates at the openmock instance
		templatesYaml := localOpenMock.ToYAML()
		response, httpErr := http.Post(openMockURL+httpPath, "application/yaml", bytes.NewReader(templatesYaml))
		if httpErr != nil {
			logrus.Errorf("Error posting templates at %s: [%s]", openMockURL, httpErr)
			return httpErr
		}
		response.Body.Close()

		logrus.Info("Posted templates!")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
}
