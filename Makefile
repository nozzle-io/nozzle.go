NOZZLE_DIR := deps/nozzle
PLOG_DIR := $(NOZZLE_DIR)/libs/plog/include
BUILD_DIR := .build

CXX := c++
AR := ar

CXXFLAGS := -std=c++17 -fno-exceptions -fno-rtti -O2

UNAME_S := $(shell uname -s)
ifeq ($(OS),Windows_NT)
	PLATFORM := windows
else ifeq ($(UNAME_S),Darwin)
	PLATFORM := macos
else
	PLATFORM := linux
endif

# Common sources
COMMON_SRCS := \
	$(NOZZLE_DIR)/src/common/ipc.cpp \
	$(NOZZLE_DIR)/src/common/registry.cpp \
	$(NOZZLE_DIR)/src/common/sender.cpp \
	$(NOZZLE_DIR)/src/common/receiver.cpp \
	$(NOZZLE_DIR)/src/common/frame.cpp \
	$(NOZZLE_DIR)/src/common/texture.cpp \
	$(NOZZLE_DIR)/src/common/device.cpp \
	$(NOZZLE_DIR)/src/common/discovery.cpp \
	$(NOZZLE_DIR)/src/common/metadata.cpp \
	$(NOZZLE_DIR)/src/common/pixel_access.cpp \
	$(NOZZLE_DIR)/src/common/channel_swizzle.cpp \
	$(NOZZLE_DIR)/src/common/format_convert.cpp \
	$(NOZZLE_DIR)/src/common/format_convert_sse2.cpp \
	$(NOZZLE_DIR)/src/common/format_convert_neon.cpp \
	$(NOZZLE_DIR)/src/common/format_resolve.cpp \
	$(NOZZLE_DIR)/src/c_api/nozzle_c.cpp \
	$(NOZZLE_DIR)/src/backends/opengl/opengl_backend.cpp

ifeq ($(PLATFORM),macos)
	CXXFLAGS += -DNOZZLE_PLATFORM_MACOS=1 -DNOZZLE_HAS_METAL=1 -DNOZZLE_HAS_OPENGL=1
	PLATFORM_SRCS := \
		$(NOZZLE_DIR)/src/backends/metal/metal_backend.mm \
		$(NOZZLE_DIR)/src/backends/metal/metal_texture.mm \
		$(NOZZLE_DIR)/src/backends/metal/metal_channel_swap.mm \
		$(NOZZLE_DIR)/src/backends/metal/metal_sync.mm \
		$(NOZZLE_DIR)/src/common/channel_swizzle_vimage.cpp \
		$(NOZZLE_DIR)/src/common/format_convert_vimage.cpp
	LDFLAGS := -framework Metal -framework IOSurface -framework Foundation -framework Accelerate -framework OpenGL -lobjc -lc++
endif

ifeq ($(PLATFORM),linux)
	CXXFLAGS += -DNOZZLE_PLATFORM_LINUX=1 -DNOZZLE_HAS_DMA_BUF=1 -DNOZZLE_HAS_OPENGL=1
	PLATFORM_SRCS := \
		$(NOZZLE_DIR)/src/backends/linux/linux_texture.cpp
	LDFLAGS := -ldrm -lgbm -lEGL -lGL -lstdc++
endif

ifeq ($(PLATFORM),windows)
	CXXFLAGS := -DNOZZLE_PLATFORM_WINDOWS=1 -DNOZZLE_HAS_D3D11=1 -DNOZZLE_HAS_OPENGL=1
	PLATFORM_SRCS := \
		$(NOZZLE_DIR)/src/backends/d3d11/d3d11_backend.cpp \
		$(NOZZLE_DIR)/src/backends/d3d11/d3d11_texture.cpp \
		$(NOZZLE_DIR)/src/backends/d3d11/d3d11_sync.cpp
	LDFLAGS := -ld3d11 -ldxgi -lopengl32 -lbcrypt -lc++
endif

INCLUDES := -I$(NOZZLE_DIR)/include -I$(NOZZLE_DIR)/src -I$(PLOG_DIR)
ifeq ($(PLATFORM),linux)
	INCLUDES += -I/usr/include/libdrm
endif

ALL_SRCS := $(COMMON_SRCS) $(PLATFORM_SRCS)
ALL_OBJS := $(patsubst %.cpp,$(BUILD_DIR)/%.o,$(patsubst %.mm,$(BUILD_DIR)/%.o,$(ALL_SRCS)))

LIB := $(BUILD_DIR)/libnozzle.a

.PHONY: all clean

all: $(LIB)

$(LIB): $(ALL_OBJS)
	@mkdir -p $(dir $@)
	$(AR) rcs $@ $^

$(BUILD_DIR)/%.o: %.cpp
	@mkdir -p $(dir $@)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

$(BUILD_DIR)/%.o: %.mm
	@mkdir -p $(dir $@)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

clean:
	rm -rf $(BUILD_DIR)
