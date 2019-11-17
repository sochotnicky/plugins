package cachet

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chat-bot/bot"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	outageReportConfig map[string][]string
	cachetAPI          string
)

// cachedComponents is Go representation of https://docs.cachethq.io/reference#get-components
type cachetComponents struct {
	Meta struct {
		Pagination struct {
			Total       int `json:"total"`
			Count       int `json:"count"`
			PerPage     int `json:"per_page"`
			CurrentPage int `json:"current_page"`
			TotalPages  int `json:"total_pages"`
			Links       struct {
				NextPage     interface{} `json:"next_page"`
				PreviousPage interface{} `json:"previous_page"`
			} `json:"links"`
		} `json:"pagination"`
	} `json:"meta"`
	Data []struct {
		ID          int           `json:"id"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Link        string        `json:"link"`
		Status      int           `json:"status"`
		Order       int           `json:"order"`
		GroupID     int           `json:"group_id"`
		Enabled     bool          `json:"enabled"`
		Meta        interface{}   `json:"meta"`
		CreatedAt   string        `json:"created_at"`
		UpdatedAt   string        `json:"updated_at"`
		DeletedAt   interface{}   `json:"deleted_at"`
		StatusName  string        `json:"status_name"`
		Tags        []interface{} `json:"tags"`
	} `json:"data"`
}

func cachetGetFailedComponents() (failed []string, err error) {
	err = nil
	url := fmt.Sprintf("%s/v1/components?status=4", cachetAPI)
	log.Printf("Getting components in outage from Cachet URL %s", url)
	resp, err := http.Get(url)

	if err != nil {
		log.Printf("Cachet API call failed: %v", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Cachet API call failed with: %d", resp.StatusCode)
		return nil, errors.New("Cachet API call failed")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed reading cachet response body: %v", err)
		return nil, err
	}
	var cc cachetComponents
	err = json.Unmarshal(body, &cc)
	if err != nil {
		log.Printf("Failed to unmarshal JSON response: %v", err)
		return nil, err
	}

	for _, failedComponent := range cc.Data {
		log.Printf("Component %s is in outage", failedComponent.Name)
		failed = append(failed, failedComponent.Name)
	}

	return
}

func checkCachet() (ret []bot.CmdResult, err error) {
	failedNames, err := cachetGetFailedComponents()
	if err != nil {
		log.Printf("Failure while getting failed components: %v", err)
		return
	}
	anyAlertConfig, anyAlert := outageReportConfig["any"]

	for _, failedService := range failedNames {
		log.Printf("Reporting alerts for %s", failedService)
		if anyAlert {
			log.Print("Reporting outages to channels with 'any' outage config")
			// Report to all channels which specified "any" outage
			for _, channel := range anyAlertConfig {
				log.Printf("Alerting about %s outage in %s", failedService, channel)
				ret = append(ret, bot.CmdResult{
					Message: fmt.Sprintf("Service '%s' is in outage as per %s",
						failedService, cachetAPI),
					Channel: channel,
				})
			}
		}

		alertConfig, found := outageReportConfig[failedService]
		if !found {
			log.Printf("No alert config for %s. Skipping", failedService)
			continue
		}
		for _, channel := range alertConfig {
			log.Printf("Alerting about %s outage in %s", failedService, channel)
			ret = append(ret, bot.CmdResult{
				Message: fmt.Sprintf("Service '%s' is in outage as per %s",
					failedService, cachetAPI),
				Channel: channel,
			})

		}
	}
	return
}

func init() {
	outageReportConfig = make(map[string][]string)
	cachetAPI = os.Getenv("CACHET_API")

	alertMaps := strings.Split(os.Getenv("CACHET_ALERT_CONFIG"), ";")
	for _, alertMap := range alertMaps {
		cc := strings.Split(alertMap, ":")
		if len(cc) > 2 {
			log.Panic("CACHET config ambiguous. Found multiple ':' in single alert")
		}
		component := cc[0]
		channel := cc[1]
		outageReportConfig[component] = append(outageReportConfig[component],
			channel)
	}

	bot.RegisterPeriodicCommandV2(
		"systemStatusCheck",
		bot.PeriodicConfig{
			CronSpec:  "@every 1m",
			CmdFuncV2: checkCachet,
		})
}
