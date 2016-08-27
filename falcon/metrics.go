package falcon

import (
	"fmt"
	"log"
	"rmqmon/funcs"
	"rmqmon/g"
	"strconv"
	"strings"
	"time"
)

var (
	OvPrefix string = "rabbitmq.overview."
	QuPrefix string = "rabbitmq.queue."
)

type MetaData struct {
	Endpoint    string      `json:"endpoint"`
	Metric      string      `json:"metric"`
	Value       interface{} `json:"value"`
	CounterType string      `json:"counterType"`
	Tags        string      `json:"tags"`
	Timestamp   int64       `json:"timestamp"`
	Step        int64       `json:"step"`
}

func NewMetric(name string, value interface{}, tags string) *MetaData {
	host := g.GetHost()
	return &MetaData{
		Metric:      name,
		Endpoint:    host,
		CounterType: fmt.Sprintf("GAUGE"),
		Tags:        tags,
		Timestamp:   time.Now().Unix(),
		Step:        g.Config().Interval,
		Value:       value,
	}
}

func (m *MetaData) String() string {
	s := fmt.Sprintf("MetaData Metric:%s Endpoint:%s Value:%v CounterType:%s Tags:%s Timestamp:%d Step:%d",
		m.Metric, m.Endpoint, m.Value, m.CounterType, m.Tags, m.Timestamp, m.Step)
	return s
}

func (m *MetaData) SetValue(v interface{}) {
	m.Value = v
}

func calcpct(l, t int64) (pct float64) {
	var err error
	if t == 0 {
		return
	}
	pct = float64(l) / float64(t) * 100.00
	if pct, err = strconv.ParseFloat(fmt.Sprintf("%.3f", pct), 64); err == nil {
		return
	}
	return
}

func qstats(s string) int64 {
	var alivequeue = g.Config().Qrunning
	for _, i := range alivequeue {
		if strings.Contains(strings.ToLower(s), i) {
			return 1
		} else {
			continue
		}
	}
	return 0
}

func aliveness(s string) int64 {
	switch s {
	case "ok":
		return 1
	default:
		return 0
	}
}

func partitions(s []string) int64 {
	lens := len(s)
	switch lens {
	case 0:
		return 1
	default:
		return 0

	}
}

func handleOverview(data []*MetaData) []*MetaData {
	ov, err := funcs.GetOverview()
	if err != nil {
		log.Println(err)
	}

	nd, err := funcs.GetNode()
	if err != nil {
		log.Println(err)
	}

	al, err := funcs.GetAlive()
	if err != nil {
		log.Println(err)
	}

	data = append(data, NewMetric(OvPrefix + "queuesTotal", ov.Queues, "")) // 队列总数
	data = append(data, NewMetric(OvPrefix + "channelsTotal", ov.Channels, ""))
	data = append(data, NewMetric(OvPrefix + "connectionsTotal", ov.Connections, ""))
	data = append(data, NewMetric(OvPrefix + "consumersTotal", ov.Consumers, ""))
	data = append(data, NewMetric(OvPrefix + "exchangesTotal", ov.Exchanges, ""))
	data = append(data, NewMetric(OvPrefix + "msgsTotal", ov.MsgsTotal, ""))
	data = append(data, NewMetric(OvPrefix + "msgsReadyTotal", ov.MsgsReadyTotal, ""))
	data = append(data, NewMetric(OvPrefix + "msgsUnackTotal", ov.MsgsUnackedTotal, ""))
	data = append(data, NewMetric(OvPrefix + "deliverTotal", ov.Deliver_get, ""))
	data = append(data, NewMetric(OvPrefix + "publishTotal", ov.Publish, ""))
	data = append(data, NewMetric(OvPrefix + "redeliverTotal", ov.Redeliver, ""))
	data = append(data, NewMetric(OvPrefix + "statsDbEvent", ov.StatsDbEvents, "")) //统计数据库事件数
	data = append(data, NewMetric(OvPrefix + "deliverRate", ov.Deliver_get_Rates.Rate, ""))
	data = append(data, NewMetric(OvPrefix + "publishRate", ov.Publish_Rates.Rate, ""))
	data = append(data, NewMetric(OvPrefix + "redeliverRate", ov.Redeliver_Rates.Rate, ""))
	data = append(data, NewMetric(OvPrefix + "ackRate", ov.Ack_Rates.Rate, ""))
	data = append(data, NewMetric(OvPrefix + "fdUsedPct", calcpct(nd.FdUsed, nd.FdTotal), ""))
	data = append(data, NewMetric(OvPrefix + "memUsedPct", calcpct(nd.MemUsed, nd.MemLimit), ""))
	data = append(data, NewMetric(OvPrefix + "socketUsedPct", calcpct(nd.SocketsUsed, nd.SocketsTotal), ""))
	data = append(data, NewMetric(OvPrefix + "erlProcsUsedPct", calcpct(nd.ErlProcUsed, nd.ErlProcTotal), "")) //消费生产比
	data = append(data, NewMetric(OvPrefix + "dpRatio", calcpct(int64(ov.Deliver_Rates.Rate), int64(ov.Publish_Rates.Rate)), ""))
	data = append(data, NewMetric(OvPrefix + "runQueue", nd.RunQueues, ""))
	data = append(data, NewMetric(OvPrefix + "isAlive", aliveness(al.Status), ""))          //读写判断
	data = append(data, NewMetric(OvPrefix + "isPartition", partitions(nd.Partitions), "")) //是否发生网络分区
	data = append(data, NewMetric(OvPrefix + "isUp", 1, ""))

	return data
}

func handleQueues(data []*MetaData) []*MetaData {
	qs, err := funcs.GetQueues()
	if err != nil {
		log.Println(err)
	}

	for _, q := range *qs {
		tags := fmt.Sprintf("name=%s,vhost=%s", q.Name, q.Vhost)
		data = append(data, NewMetric(QuPrefix + "messages", q.Messages, tags))
		data = append(data, NewMetric(QuPrefix + "messages_ready", q.MessagesReady, tags))
		data = append(data, NewMetric(QuPrefix + "messages_unacked", q.MessagesUnacked, tags))
		data = append(data, NewMetric(QuPrefix + "deliver_get", q.Deliver_get.Rate, tags))
		data = append(data, NewMetric(QuPrefix + "publish", q.Publish.Rate, tags))
		data = append(data, NewMetric(QuPrefix + "redeliver", q.Redeliver.Rate, tags))
		data = append(data, NewMetric(QuPrefix + "ack", q.Ack.Rate, tags))
		data = append(data, NewMetric(QuPrefix + "memory", q.Memory, tags))
		data = append(data, NewMetric(QuPrefix + "consumers", q.Consumers, tags))
		data = append(data, NewMetric(QuPrefix + "consumer_utilisation", q.ConsumerUtil * 100.00, tags))
		data = append(data, NewMetric(QuPrefix + "status", qstats(q.Status), tags))
		data = append(data, NewMetric(QuPrefix + "dpratio", calcpct(int64(q.Deliver_get.Rate), int64(q.Publish.Rate)), tags))
	}

	return data
}

func handleSickRabbit(data []*MetaData) []*MetaData {
	data = append(data, NewMetric(OvPrefix + "isUp", 0, ""))
	return data
}

func Collector() {
	m := make([]*MetaData, 0)

	if !funcs.CheckAlive() {
		log.Println("ERROR: Can not connect to rabbit.")
		m = handleSickRabbit(m)
	} else {
		m = handleOverview(m)
		m = handleQueues(m)
	}

	log.Printf("Send to %s, size: %d.", g.Config().Falcon.Api, len(m))
	// log for debug
	if g.Config().Debug {
		for _, m := range m {
			log.Println(m.String())
		}
	}

	sendDatas(m)
}
