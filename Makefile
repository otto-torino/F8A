NAME    := f8a
BIN     := build/$(NAME)
PKG_DIR := $(NAME)
TARBALL := $(NAME).tar.xz
DESKTOP := $(PKG_DIR)/usr/local/share/applications/$(NAME).desktop
# Display name shown in app lists, read from the Go source so it cannot drift
# from the window title.
TITLE   := $(shell sed -n 's/^const AppTitle = "\(.*\)"$$/\1/p' utils/appinfo.go)

.PHONY: build package user-install user-uninstall clean

build:
	go build -o $(BIN) .

# Build the distribution tarball. `fyne package` uses --name for both the file
# names and the Name= shown in app lists, and emits a Makefile that installs
# the icon outside the hicolor theme, so the tarball is unpacked, fixed and
# repacked.
package: build
	test -n "$(TITLE)" || { echo "AppTitle not found in utils/appinfo.go"; exit 1; }
	rm -rf $(PKG_DIR) $(TARBALL)
	fyne package -os linux --name $(NAME) --exe $(BIN) --icon Icon.png
	tar xf $(TARBALL)
	sed -i 's/^Name=.*/Name=$(TITLE)/' $(DESKTOP)
	cp packaging/Makefile $(PKG_DIR)/Makefile
	rm -f $(TARBALL)
	tar cJf $(TARBALL) $(PKG_DIR)

user-install: package
	$(MAKE) -C $(PKG_DIR) user-install

user-uninstall:
	$(MAKE) -C $(PKG_DIR) user-uninstall

clean:
	rm -rf build $(PKG_DIR)
