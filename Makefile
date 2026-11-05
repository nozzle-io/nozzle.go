NOZZLE_DIR := deps/nozzle
PLOG_DIR := $(NOZZLE_DIR)/libs/plog/include
BUILD_DIR := .build

CXX := c++
AR := ar

CXXFLAGS := -std=c++17 -fno-exceptions -fno-rtti -O2
EXTRA_CXXFLAGS ?=

UNAME_S := $(shell uname -s)
ifeq ($(OS),Windows_NT)
	PLATFORM := windows
else ifeq ($(UNAME_S),Darwin)
	PLATFORM := macos
else
	PLATFORM := linux
endif

# Common sources
include $(NOZZLE_DIR)/sources/sources.mk

COMMON_SRCS := $(NOZZLE_COMMON_SRCS) $(NOZZLE_OPENGL_SRCS)

ifeq ($(PLATFORM),macos)
	CXXFLAGS += -DNOZZLE_PLATFORM_MACOS=1 -DNOZZLE_HAS_METAL=1 -DNOZZLE_HAS_OPENGL=1
	PLATFORM_SRCS := $(NOZZLE_MACOS_SRCS)
	LDFLAGS := -framework Metal -framework IOSurface -framework Foundation -framework Accelerate -framework OpenGL -lobjc -lc++
endif

ifeq ($(PLATFORM),linux)
	CXXFLAGS += -DNOZZLE_PLATFORM_LINUX=1 -DNOZZLE_HAS_DMA_BUF=1 -DNOZZLE_HAS_OPENGL=1
	PLATFORM_SRCS := $(NOZZLE_LINUX_SRCS)
	LDFLAGS := -ldrm -lgbm -lEGL -lGL -lstdc++
endif

ifeq ($(PLATFORM),windows)
	CXXFLAGS := -DNOZZLE_PLATFORM_WINDOWS=1 -DNOZZLE_HAS_D3D11=1 -DNOZZLE_HAS_OPENGL=1
	PLATFORM_SRCS := $(NOZZLE_WINDOWS_SRCS)
	LDFLAGS := -ld3d11 -ldxgi -lopengl32 -lbcrypt -lc++
endif

CXXFLAGS += $(EXTRA_CXXFLAGS)

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
