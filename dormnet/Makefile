include $(TOPDIR)/rules.mk

PKG_NAME:=dormnet
PKG_VERSION:=0.1.0
PKG_RELEASE:=2

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)
DORMNET_SRC_DIR?=$(realpath $(dir $(lastword $(MAKEFILE_LIST)))/../dormnet-src)

PKG_LICENSE:=AGPL3.0
PKG_MAINTAINER:=Haven Madray <sgpublic2002@gmail.com>

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_BUILD_FLAGS:=no-mips16

PKG_BUILD_TIME:=$(shell date -u -Iseconds)

GO_PKG:=github.com/openwrt-dormnet/dormnet
GO_PKG_LDFLAGS_X:=$(GO_PKG)/internal.constants.version=$(PKG_VERSION) $(GO_PKG)/internal.constants.buildTime=$(PKG_BUILD_TIME)
GO_PKG_LDFLAGS:=-s -w

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/dormnet
  SECTION:=net
  CATEGORY:=Network
  TITLE:=DormNet
  URL:=https://github.com/openwrt-dormnet/dormnet
  DEPENDS:=$(GO_ARCH_DEPENDS) +luci-app-mwan3
  PROVIDES:=dormnet
endef

define Package/dormnet/description
  Manage your campus network with DormNet on OpenWrt.
endef

define Package/dormnet/conffiles
/etc/config/dormnet
endef

define Build/Prepare
	$(RM) -r $(PKG_BUILD_DIR)
	$(INSTALL_DIR) $(PKG_BUILD_DIR)
	$(CP) $(DORMNET_SRC_DIR)/* $(PKG_BUILD_DIR)/
endef

define Package/dormnet/install
	$(call GoPackage/Package/Install/Bin,$(1))

	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_INSTALL_DIR)/usr/bin/dormnet $(1)/usr/bin/dormnet
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./root/etc/init.d/dormnet $(1)/etc/init.d
endef

$(eval $(call GoBinPackage,dormnet))
$(eval $(call BuildPackage,dormnet))
