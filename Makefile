BINARY := gopener
BUILD_DIR := build
INSTALL_DIR := $(HOME)/.local/bin
CMD := ./cmd/gopener

.PHONY: all build install clean

all: build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY)"

clean:
	rm -rf $(BUILD_DIR)
