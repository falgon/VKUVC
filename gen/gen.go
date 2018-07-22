package gen

import (
	"../vip/detali"
	"bytes"
	"errors"
	"flag"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"text/template"
)

var (
	Region                = flag.String("region", "ap-northeast-1", "Region name")
	redudancyInstancesTag = flag.String("redInstanceTag", "", "Tag name indicating that it is a member of the redundant configration. Ex. Key=foo,Value=bar (single only)")
	argInput              = flag.String("input", "", "input template file")
	vIPKey                = flag.String("VIPKey", "", "Key of tag indicating redundant VIP")
	rtable                = flag.String("rtable", "", "Route table id")
)

type Member struct {
	State        string
	UnicastPeer  []string
	VIP          string
	NotifyMaster string
}

// 可用構成に含まれる Key=VIP, Value=VIP の Value を取得,
// 実行したインスタンスに VIP が設定されているかを真偽で,
// また可用構成に含まれる実行したインスタンス以外のプライベート IP アドレスを, エラー情報とともに返す.
func getVIP(svc *ec2.EC2) (vipval string, isVip bool, red []string, err error) {
	var redudancies *ec2.DescribeInstancesOutput
	isVip = false
	if redudancies, err = detail.GetRedudancies(svc, redudancyInstancesTag); !detail.OutError(err) {
		err = errors.New("Failed to get Redudancies")
		return
	}
	var instance string
	if instance, err = detail.GetThisInstanceId(); !detail.OutError(err) {
		return
	}
	red = make([]string, 0)

	for _, t := range redudancies.Reservations {
		for _, i := range t.Instances {
			for _, tag := range i.Tags {
				if *tag.Key == *vIPKey {
					vipval = *tag.Value
					if instance == *i.InstanceId {
						isVip = true
					}
				}
			}
			if instance != *i.InstanceId {
				red = append(red, *i.PrivateIpAddress)
			}
		}
	}
	if vipval == "" {
		err = errors.New("The tag of VIP parse error.\nProbably the VIP tag is not set correctly.")
		return
	}
	return
}

func CheckArg() {
	t1, t2 := "At least one ", " must be specified."
	var message string
	if *argInput == "" {
		message = "template file"
	} else if *redudancyInstancesTag == "" {
		message = "tag"
	} else if *vIPKey == "" {
		message = "the Key of VIP"
	} else if *rtable == "" {
		message = "the route table ID"
	}
	if message != "" {
		detail.OutError(errors.New(t1 + message + t2))
		os.Exit(1)
	}
}

func GetAlived(svc *ec2.EC2) (vipvalue, state string, red []string, err error) {
	var isVip bool
	if vipvalue, isVip, red, err = getVIP(svc); err != nil {
		return
	}
	if isVip {
		state = "MASTER"
	} else {
		state = "BACKUP"
	}
	return
}

func GenerateConfig(vipvalue, state *string, red *[]string) {
	t := template.Must(template.ParseFiles(*argInput))
	if _, value, err := detail.ParseKeyValue(redudancyInstancesTag); !detail.OutError(err) {
		os.Exit(1)
	} else {
		mem := Member{
			*state,
			*red,
			*vipvalue,
			"notify_master \"/etc/keepalived/notify_master -region='" + *Region +
				"' -dcidr='" + *vipvalue +
				"' -rtable='" + *rtable +
				"' -viptag='Key=" + *vIPKey + ",Value=" + *vipvalue +
				"' -redInstanceTag='Key=Name,Value=" + value + "'\"",
		}
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, mem); !detail.OutError(err) {
			os.Exit(1)
		}
		if f, err := os.Create("keepalived.conf"); !detail.OutError(err) {
			os.Exit(1)
		} else {
			defer f.Close()
			f.WriteString(tpl.String())
		}
	}
}
