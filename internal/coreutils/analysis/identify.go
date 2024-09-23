package analysis

import (
	"fmt"

	pools "github.com/PlagueByteSec/sdakit-project/v2/internal/datapools"
	"github.com/PlagueByteSec/sdakit-project/v2/internal/requests"
	"github.com/PlagueByteSec/sdakit-project/v2/internal/shared"
	"github.com/fhAnso/astkit"
)

func (check *SubdomainCheck) MailServer() {
	if requests.DnsIsMX(shared.GDnsResolver, check.Subdomain) {
		check.ConsoleOutput <- " | + Mail Server "
		if check.isExchange() {
			check.ConsoleOutput <- "(Microsoft Exchange)\n"
		} else {
			check.ConsoleOutput <- "\n"
		}
		pools.ManagePool(pools.PoolAction(pools.PoolAppend), check.Subdomain, &shared.GPoolBase.PoolMailSubdomains)
	}
}

func (check *SubdomainCheck) api() {
	url := astkit.MakeUrl(astkit.HTTP(astkit.Basic), check.Subdomain)
	for idx := 0; idx < len(methods); idx++ {
		response := check.AnalysisSendRequest(AnalysisRequestConfig{Method: methods[idx], URL: url, Header: "", Value: ""})
		if response == nil {
			continue
		}
		statusCode := response.StatusCode
		if cloudflareError(statusCode, check.Subdomain) {
			continue
		}
		score, info := check.isPossibleApi(response)
		if score != 0 {
			pools.ManagePool(pools.PoolAction(pools.PoolAppend), check.Subdomain, &shared.GPoolBase.PoolApiSubdomains)
			check.ConsoleOutput <- fmt.Sprintf(" | + API [SCORE:%d] (%s: %s)\n", score, methods[idx], info)
			break
		}
	}
}

func (check *SubdomainCheck) investigateHtmlResponse() {
	url := astkit.MakeUrl(astkit.HTTP(astkit.Basic), check.Subdomain)
	response := check.getResponse(url) // GET
	if response == nil {
		return
	}
	defer response.Body.Close()
	body := string(check.responseGetBody(response))
	if len(body) == 0 {
		return
	}
	check.checkPage("login", detectLogin, body)
	check.checkPage("cms", detectCMS, body)
}
