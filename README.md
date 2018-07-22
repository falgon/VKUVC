# VPC-EC2 + Keepalived Utilities = VRRP on Cloud

[![Build Status](https://travis-ci.org/falgon/VKUVC.svg?branch=master)](https://travis-ci.org/falgon/VKUVC)

The Utility for achieving availability using Keepalived with VPC-EC2. Build:

```sh
$ git clone https://github.com/falgon/VKUVC && cd VKUVC && make get && make
```

## Demo video

[![Youtube](http://img.youtube.com/vi/b6y3CnaTiG0/hqdefault.jpg)](https://www.youtube.com/watch?v=b6y3CnaTiG0)

## Usage

This tool is based on the contents of the following items.

* Perform redundancy with AWS EC2(At least two EC2 instances of normal system and standby system are necessary).
* Routing setting to EC2 instance of normal system is completed with EC2-VPC.

### ./dst/keepalived\_configure

The tool to automatically generate keepalived.conf. The flags are:
```sh
$ ./dst/keepalived_configure --help
Usage of ./dst/keepalived_configure:
  -VIPKey string
        Key of tag indicating redundant VIP
  -input string
        input template file
  -redInstanceTag string
        Tag name indicating that it is a member of the redundant configration. Ex. Key=foo,Value=bar (single only)
  -region string
        Region name (default "ap-northeast-1")
  -rtable string
        Route table id
```
Also, configure will load the template as follows. E.g.:
```sh
$ ./dst/keepalived_configure\
	-region="ap-northeast-1"\
	-redInstanceTag="Key=Name,Value=RedudancySystem"\
	-VIPKey="VIP"\
	-rtable="rtb-xxxxxxxx"\
	-input="keepalived.conf.template"
```
The [Sample template](https://github.com/falgon/VKUVC/tree/master/keepalived.conf.template) is as follows.
```sh
vrrp_instance AWS {
    state {{.State}}
    nopreempt
    interface eth0

    unicast_peer {
    {{range $var := .UnicastPeer}}
        {{$var}}
    {{end}}
    }

    lvs_sync_daemon_interface eth0
    virtual_router_id 100
    priority 100
    advert_int 1

    authentication {
        auth_type PASS
        auth_pass aliv
    }

    virtual_ipaddress {
    {{.VIP}}
    }

    {{.NotifyMaster}}
}
```

### ./dst/notify\_master

If you use it with the previous configure, just put `./dst/notify_master` under `/etc/keepalived`. Flags:
```sh
% ./dst/notify_master --help                 (18-07-23)[5:01:34]
Usage of ./dst/notify_master:
  -dcidr string
        destination cidr block value (default "192.168.1.1/32")
  -redInstanceTag string
        Tag name indicating that it is a member of a redundant configuration. To set this flag, the tagvip flag must also be specified.
        Ex. Key=foo,Value=bar (Single only)
  -region string
        Region name (default "ap-northeast-1")
  -rtable string
        Route table
  -viptag string
        Tag names too be added to the time of route creation. To set this flag, the redudancyInstancesTag flag must also be specified.
        Ex. Key=foo,Value=bar Key=hoge,Value=piyo
```
