package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete specified template set key on remote instance",
	Long:  "delete specified template set key on remote instance",
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

		// verify setKey was set
		if setKey == "" {
			return fmt.Errorf("Must use -k with delete")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		httpPath := openMockURL + "/api/v1/template_sets/" + setKey

		client := &http.Client{}
		req, err := http.NewRequest("DELETE", httpPath, nil)
		if err != nil {
			logrus.Errorf("Error deleting templates at %s / %s: [%s]", openMockURL, setKey, err)
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			logrus.Errorf("Error deleting templates at %s / %s: [%s]", openMockURL, setKey, err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 204 {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Errorf("Error deleting templates at %s / %s: [%s]", openMockURL, setKey, err)
				return err
			}
			logrus.Errorf("Error deleting templates at %s / %s: [%s]", openMockURL, setKey, respBody)
			return fmt.Errorf("")
		}

		logrus.Info("Deleted templates!")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
