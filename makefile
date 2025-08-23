BINARY_NAME=barghman

CONFIG_PATH=$(HOME)/.config/barghman
SYSTEMD_PATH=$(HOME)/.config/systemd/user
CACHE_PATH=$(HOME)/.cache/barghman
BIN_PATH=$(HOME)/.local/bin

.PHONY: build
build:
	go build -o $(BINARY_NAME) ./...

.PHONY: install
install: build

	mkdir -p $(CONFIG_PATH)
	mkdir -p $(SYSTEMD_PATH)
	mkdir -p $(CACHE_PATH)
	mkdir -p $(BIN_PATH)

	cp $(BINARY_NAME) $(BIN_PATH)
	chmod +x $(BIN_PATH)/$(BINARY_NAME)

	@if [ ! -f $(CONFIG_PATH)/config.toml ]; then \
		echo "Config file not exists, creating it from example"; \
		cp example.toml $(CONFIG_PATH)/config.toml; \
	else \
		echo "Config exists, skipping: $(CONFIG_PATH)/config.toml"; \
	fi

	@sed -e "s|{{INSTALL_PATH}}|$(BIN_PATH)|g" \
	     -e "s|{{CONFIG_PATH}}|$(CONFIG_PATH)|g" \
	     -e "s|{{CACHE_PATH}}|$(CACHE_PATH)|g" \
	     systemd/$(BINARY_NAME).service.template > $(SYSTEMD_PATH)/$(BINARY_NAME).service

	systemctl --user daemon-reload
	systemctl --user enable $(BINARY_NAME).service

.PHONY: uninstall
uninstall:
	systemctl --user disable $(BINARY_NAME).service || true
	rm -f $(SYSTEMD_PATH)/$(BINARY_NAME).service
	rm -f $(BIN_PATH)/$(BINARY_NAME)
	rm -rf $(CONFIG_PATH)
	rm -rf $(CACHE_PATH)
	rm -f $(BINARY_NAME)

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(CACHE_PATH)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  install   - Build and install the service"
	@echo "  uninstall - Remove the service and binary"
	@echo "  clean     - Remove build artifacts and cache"