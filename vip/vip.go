package vip

import (
	"./detali"
	"./tags"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
)

var (
	region                = flag.String("region", "ap-northeast-1", "Region name")
	dcidr                 = flag.String("dcidr", "192.168.1.1/32", "destination cidr block value")
	rtable                = flag.String("rtable", "", "Route table")
	tagvip                = flag.String("viptag", "", "Tag names too be added to the time of route creation. To set this flag, the redudancyInstancesTag flag must also be specified.\nEx. Key=foo,Value=bar Key=hoge,Value=piyo")
	redudancyInstancesTag = flag.String("redInstanceTag", "", "Tag name indicating that it is a member of a redundant configuration. To set this flag, the tagvip flag must also be specified.\nEx. Key=foo,Value=bar (Single only)")
)

// * svc: 初期化済みの *ec2.EC2
//
// * destinationCidrBlock: VIP
//
// * routeTableId: ルートテーブル ID
//
// 実行したインスタンスへのルーティングの設定を行う.
// インスタンス ID の取得に失敗した場合, `nil, err` を, そうでない場合 `CreateRoute` を実行した結果をそのまま返す.
func createRoute(svc *ec2.EC2, destinationCidrBlock, routeTableId string) (*ec2.CreateRouteOutput, error) {
	var (
		instanceId string
		err        error
	)

	if instanceId, err = detail.GetThisInstanceId(); err != nil {
		return nil, err
	}
	input := &ec2.CreateRouteInput{
		InstanceId:           aws.String(instanceId),
		DestinationCidrBlock: aws.String(destinationCidrBlock),
		RouteTableId:         aws.String(routeTableId),
	}
	return svc.CreateRoute(input)
}

// * svc: 初期化済みの *ec2.EC2
//
// * destinationCidrBlock: VIP
//
// * routeTableId: ルートテーブル ID
//
// 対象となるルーティング設定を削除する.
// `DeleteRoute` を実行した結果をそのまま返す.
func deleteRoute(svc *ec2.EC2, destinationCidrBlock, routeTableId string) (*ec2.DeleteRouteOutput, error) {
	input := &ec2.DeleteRouteInput{
		DestinationCidrBlock: aws.String(destinationCidrBlock),
		RouteTableId:         aws.String(routeTableId),
	}
	return svc.DeleteRoute(input)
}

// * svc: 初期化済みの *ec2.EC2
//
// redudancyInstancesTag で指定された冗長構成のうち,
// 実行したインスタンス以外に VIP タグが付与されていた場合,
// それを削除する.
func deleteOtherVIPTag(svc *ec2.EC2) bool {
	if *redudancyInstancesTag != "" {
		if redudancies, err := detail.GetRedudancies(svc, redudancyInstancesTag); !detail.OutError(err) {
			return false
		} else {
			var key, value string
			if key, value, err = detail.ParseKeyValue(tagvip); !detail.OutError(err) {
				return false
			}
			var instanceId string
			if instanceId, err = detail.GetThisInstanceId(); !detail.OutError(err) {
				return false
			}

			if vipTag, err := tags.GenerateTags(*tagvip); !detail.OutError(err) {
				return false
			} else {
				for _, t := range redudancies.Reservations {
					for _, i := range t.Instances {
						for _, tag := range i.Tags {
							if *tag.Key == key && *tag.Value == value && instanceId != *i.InstanceId {
								fmt.Println("Found other VIP tag. Its instance id is <" + *i.InstanceId + ">")
								fmt.Println("Removing other VIP tag...")
								if err = tags.DeleteTag(svc, *i.InstanceId, vipTag); !detail.OutError(err) {
									return false
								}
								fmt.Println("Remove successful")
							}
						}
					}
				}
			}
		}
	}
	return true
}

func giveVIPTag(svc *ec2.EC2) bool {
	if *tagvip != "" {
		if tag, err := tags.GenerateTags(*tagvip); !detail.OutError(err) {
			return false
		} else {
			var instanceId string
			if instanceId, err = detail.GetThisInstanceId(); !detail.OutError(err) {
				return false
			}
			if err = tags.CreateTag(svc, instanceId, tag); !detail.OutError(err) {
				return false
			}
		}
	}
	return true
}

// * reg: レジオン
//
// * destinationCidrBlock: VIP
//
// * routeTableId: 対象となるルートテーブル ID
//
// * redudancyInstancesTag: 冗長構成が組まれている EC2 インスタンスに付与されているタグ
//
// ルートテーブルから既存の VIP を削除し, 新たに VIP を設定する.
// タグが指定されていた場合は, 実行したインスタンスにタグを付与する.
// さらに, 2 つ以上のインスタンスに固有のタグ(つまり, 冗長構成の一員であることを示すタグ)が設定されていて,
// かつその設定されているインスタンスの中で, 実行したインスタンスと異なるインスタンスに VIP タグが設定されている場合,
// そのインスタンスから VIP のタグを削除する.
//
// 予め VIP を冗長構成のうちのただ 1 つ, ルーティングテーブルに設定しておかなければならない.
func SetRoute(reg, destinationCidrBlock, routeTableId string) {
	cfg := aws.Config{
		Region: aws.String(reg),
	}
	svc := ec2.New(session.New(&cfg))

	fmt.Println("Clear the existing route table...")
	if _, err := deleteRoute(svc, destinationCidrBlock, routeTableId); !detail.OutError(err) {
		return
	} else {
		fmt.Println("Successful")
	}

	fmt.Println("Add new route table...")
	if _, err := createRoute(svc, destinationCidrBlock, routeTableId); !detail.OutError(err) {
		return
	} else {
		fmt.Println("Successful")
		if !giveVIPTag(svc) {
			return
		}
		if !deleteOtherVIPTag(svc) {
			return
		}
	}
}

// コマンドライン引数をチェックし, `SetRoute` を実行する.
func SetRouteWithArgCheck() {
	if *rtable == "" {
		fmt.Println("At least one destination Cidir table")
		os.Exit(1)
	} else if *tagvip != "" || *redudancyInstancesTag != "" { // VIP タグ指定があるなら, 冗長構成の ec2 インスタンスにつけられているタグを指定しなければならない.
		if *redudancyInstancesTag == "" {
			fmt.Println("At least one redInstanceTag")
			os.Exit(1)
		} else if *tagvip == "" {
			fmt.Println("At least one viptag")
			os.Exit(1)
		}
	}

	SetRoute(*region, *dcidr, *rtable)
}
