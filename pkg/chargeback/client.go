package chargeback

const (
	Group = "chargeback.coreos.com"
	TPRVersion = "prealpha"
)

type QueryInterface interface {
	RESTClient() rest.Interface

}