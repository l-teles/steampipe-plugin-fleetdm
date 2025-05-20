STEAMPIPE_INSTALL_DIR ?= ~/.steampipe
BUILD_TAGS = netgo
install:
	go build -o $(STEAMPIPE_INSTALL_DIR)/plugins/hub.steampipe.io/plugins/l-teles/fleetdm@latest/steampipe-plugin-fleetdm.plugin -tags "${BUILD_TAGS}" *.go
	
# LOCAL DEVELOPMENT
# install:
# 	go build -o ~/.steampipe/plugins/local/fleetdm/fleetdm.plugin *.go

# Linux ARM64
# install:
# 	GOOS=linux GOARCH=arm64 go build -o ~/.steampipe/plugins/local/fleetdm/fleetdm-linux.plugin *.go