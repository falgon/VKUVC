package tags

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

func splitxs(instances string, s string) (r []*string) {
	for _, i := range strings.Split(instances, s) {
		r = append(r, aws.String(i))
	}
	return
}

// * cli: 初期化済みの *ec2.EC2
//
// * instances: (単数 ∈)複数のインスタンス情報の文字列
//
// * tags: 付与するタグ
//
// 指定した EC2 リソースに 1 つ以上のタグを追加または上書きし, エラー情報を返す.
// 各リソースは、最大 50 のタグを持つことができ, 各タグはキーとオプションの値で構成されており,
// タグキーはリソースごとに一意でなければならない.
//
// 使用 API:
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateTags>
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#CreateTagsInput>
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#CreateTagsOutput>
func CreateTag(cli *ec2.EC2, instances string, tags []*ec2.Tag) (err error) {
	input := &ec2.CreateTagsInput{
		Resources: splitxs(instances, ","),
		Tags:      tags,
	}
	_, err = cli.CreateTags(input)
	return
}

// * tags: タグ情報の文字列
//
// スペース区切りのタグ文字列(Ex. 'Key=foo',Value=bar Key=hoge, Value=piyo')からそれぞれ ec2.Tag を生成し
// それらのスライスとエラー情報を返す.
//
// 使用 API:
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#Tag>
func GenerateTags(tags string) (r []*ec2.Tag, err error) {
	for _, tag := range strings.Split(tags, " ") {
		var tagKey, tagValue string
		for _, t := range strings.Split(tag, ",") {
			val := strings.Split(t, "=")
			switch val[0] {
			case "Key":
				tagKey = val[1]
			case "Value":
				tagValue = val[1]
			default:
				err = errors.New("Tags parse error.")
			}
		}
		Tag := &ec2.Tag{
			Key:   aws.String(tagKey),
			Value: aws.String(tagValue),
		}
		r = append(r, Tag)
	}
	return
}

// * cli: 初期化済みの *ec2.EC2
//
// * instances: (単数 ∈)複数のインスタンス情報の文字列
//
// * tags: 削除対象となるタグ
//
// 指定した EC2 リソースから 1 つ以上のタグを除き, エラー情報を返す.
//
// 使用 API:
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DeleteTags>
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DeleteTagsInput>
//
// * <https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DeleteTagsOutput>
func DeleteTag(cli *ec2.EC2, instances string, tags []*ec2.Tag) (err error) {
	input := &ec2.DeleteTagsInput{
		Resources: splitxs(instances, ","),
		Tags:      tags,
	}
	_, err = cli.DeleteTags(input)
	return
}
