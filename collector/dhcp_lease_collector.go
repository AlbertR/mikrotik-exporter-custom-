package collector

//import "github.com/prometheus/client_golang/prometheus"
//
//type dhcpLeaseCollector struct {
//	props        []string
//	descriptions *prometheus.Desc
//}
//
//func (c *dhcpLeaseCollector) init() {
//	c.props = []string{"active-mac-address", "status", "expires-after", "active-address", "host-name"}
//
//	labelNames := []string{"name", "address", "activemacaddress", "status", "expiresafter", "activeaddress", "hostname"}
//	c.descriptions = description("dhcp", "leases_metrics", "number of metrics", labelNames)
//}
//
//func newDHCPLCollector() routerOSCollector {
//
//}
