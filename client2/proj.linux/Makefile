EXECUTABLE = minispace

INCLUDES = -I.. -I../Classes

SOURCES = main.cpp \
        ../Classes/AppDelegate.cpp \
	../Classes/net_node.cpp \
	../Classes/jsoncpp.cpp \
	../Classes/role.cpp \
	../Classes/login_scene.cpp \
	../Classes/game_scene.cpp

#COCOS_ROOT = /home/lijie/projects/cocos2d-x
#export COCOS_ROOT
include $(COCOS_ROOT)/cocos2dx/proj.linux/cocos2dx.mk

INCLUDES += -I$(COCOS_ROOT)/extensions
INCLUDES += -I$(COCOS_ROOT)/external/chipmunk/include/chipmunk
$(warning $(INCLUDES))

SHAREDLIBS += -lcocos2d
STATICLIBS += $(LIB_DIR)/libextension.a
COCOS_LIBS = $(LIB_DIR)/libcocos2d.so

$(TARGET): $(OBJECTS) $(STATICLIBS) $(COCOS_LIBS) $(CORE_MAKEFILE_LIST)
	@mkdir -p $(@D)
	$(LOG_LINK)$(CXX) $(CXXFLAGS) $(OBJECTS) -o $@ $(SHAREDLIBS) $(STATICLIBS) -lpthread -lwebsockets -lGL -lz

$(OBJ_DIR)/%.o: %.cpp $(CORE_MAKEFILE_LIST)
	@mkdir -p $(@D)
	$(LOG_CXX)$(CXX) $(CXXFLAGS) $(INCLUDES) $(DEFINES) $(VISIBILITY) -c $< -o $@

$(OBJ_DIR)/%.o: ../%.cpp $(CORE_MAKEFILE_LIST)
	@mkdir -p $(@D)
	$(LOG_CXX)$(CXX) $(CXXFLAGS) $(INCLUDES) $(DEFINES) $(VISIBILITY) -c $< -o $@
