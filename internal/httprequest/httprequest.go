package httprequest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Request interface {
	RequestGet(url string) (rsp *http.Response, err error)
	RequestPost(orderNum string) (err error)
}

func NewHTTPRequst(BaseURL *url.URL) *HTTPRequest { //, httpClient *http.Client, CalcSys string) {
	return &HTTPRequest{
		BaseURL:    BaseURL,
		httpClient: http.DefaultClient,
	}
}

// структура для внешних запросов
type HTTPRequest struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

func (cl *HTTPRequest) RequestPost(orderNum string) (err error) {
	// создание JSON для запроса в систему начисления баллов
	bodyJSON := fmt.Sprintf("{\"order\":\"%s\"}", orderNum)
	// запрос регистрации заказа в системе расчета баллов
	rPost, err := cl.httpClient.Post(cl.BaseURL.String(), "application/json", strings.NewReader(bodyJSON))
	if err != nil {
		log.Printf("http Post request in ServiceNewOrderLoad error:%s", err)
	}
	// проверяем статус ответа внешнего сервиса
	if rPost.StatusCode != http.StatusAccepted {
		log.Printf("http POST request has wromg status: %s", rPost.Status)
	}
	// освобождаем ресурс
	defer rPost.Body.Close()
	return err
}

func (cl *HTTPRequest) RequestGet(orderNum string) (rsp *http.Response, err error) {
	// создаем линк обновления статуса заказов в внешнем сервисе
	linkUpd := fmt.Sprintf("%s/%s", cl.BaseURL.String(), orderNum)
	// делаем запрос во внешний сервис
	rsp, err = http.Get(linkUpd)
	if err != nil {
		log.Printf("remoute service error: %s", err)
	}
	return rsp, err
}
