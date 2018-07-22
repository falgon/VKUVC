GO=go
GOGET=$(GO) get
BUILD=$(GO) build
DST=./dst
OUTPUTQ1=notify_master
OUTPUTQ3=keepalived_configure
BIN1=$(DST)/$(OUTPUTQ1)
BIN3=$(DST)/$(OUTPUTQ3)

all: build

build:
	@mkdir -p dst
	$(BUILD) -o $(OUTPUTQ1) ./notify_master.go
	$(BUILD) -o $(OUTPUTQ3) ./keepalived_configure.go
	@mv $(OUTPUTQ1) $(OUTPUTQ3) dst
	@echo "Bins are in $(DST) :)"

get:
	$(GOGET) github.com/aws/aws-sdk-go/aws\
		github.com/aws/aws-sdk-go/aws/awserr\
	       	github.com/aws/aws-sdk-go/aws/session\
		github.com/aws/aws-sdk-go/service/ec2

run_notify_master:
	$(BIN1)\
		-region="region"\
		-dcidr="virtual ip address"\
	       	-rtable="route table id"\
	       	-viptag="Key=VIP,Value=virtual ip address"\
	       	-redInstanceTag="Key=key,Value=RedudancySystem"

run_q3:
	$(BIN3)\
	       	-input="keepalived.conf.template"\
	       	-redInstanceTag="Key=key,Value=RedudancySystem"\
		-VIPKey="VIP"\
		-rtable="route table id"

clean:
	@$(RM) -rf dst
