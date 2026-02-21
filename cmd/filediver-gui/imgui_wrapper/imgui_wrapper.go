// Package imgui_wrapper provides a basic framework for an IMGui app.
package imgui_wrapper

/*
#cgo windows LDFLAGS: -ldwmapi

// GLFW
typedef struct GLFWwindow GLFWwindow;
typedef struct GLFWmonitor GLFWmonitor;
typedef struct GLFWvidmode {
	int width;
	int height;
	int redBits;
	int greenBits;
	int blueBits;
	int refreshRate;
} GLFWvidmode;
typedef void (*GLFWwindowsizefun)(GLFWwindow* window, int width, int height);
typedef void (*GLFWwindowrefreshfun)(GLFWwindow* window);
int glfwWindowShouldClose(GLFWwindow *window);
void glfwPollEvents();
void glfwGetFramebufferSize(GLFWwindow *window, int *width, int *height);
void glfwSwapInterval(int interval);
GLFWwindow *glfwGetCurrentContext();
void glfwMakeContextCurrent(GLFWwindow *window);
void *glfwGetWindowUserPointer(GLFWwindow *window);
void glfwSetWindowUserPointer(GLFWwindow *window, void *pointer);
GLFWmonitor *glfwGetPrimaryMonitor();
GLFWvidmode *glfwGetVideoMode(GLFWmonitor *monitor);
void glfwSwapBuffers(GLFWwindow *window);
void glfwDestroyWindow(GLFWwindow *window);
void glfwTerminate();
GLFWwindowsizefun glfwSetWindowSizeCallback(GLFWwindow *window, GLFWwindowsizefun callback);
GLFWwindowrefreshfun glfwSetWindowRefreshCallback(GLFWwindow* window, GLFWwindowrefreshfun callback);

// cimgui
typedef struct ImGuiContext ImGuiContext;
typedef struct ImDrawData ImDrawData;
void igNewFrame();
ImDrawData* igGetDrawData();

// ImGui implementation
void ImGui_ImplOpenGL3_NewFrame();
void ImGui_ImplGlfw_NewFrame();
void ImGui_ImplOpenGL3_RenderDrawData(ImDrawData* draw_data);
void ImGui_ImplOpenGL3_Shutdown();
void ImGui_ImplGlfw_Shutdown();

// cimgui-go C++ wrapper stuff
typedef void (*VoidCallback)();
void glfw_render(GLFWwindow *window, VoidCallback renderLoop);

// custom functions
void goWindowResizeCallback(GLFWwindow* window, int height, int width);
void goWindowRefreshCallback(GLFWwindow* window);

// Windows dark window decoration
#ifdef _WIN32
#define DWMWA_USE_IMMERSIVE_DARK_MODE 20
// We're using the non-typedef'd names for windows data types here, see
// https://learn.microsoft.com/en-us/windows/win32/winprog/windows-data-types.
void *glfwGetWin32Window(GLFWwindow *window);
long DwmSetWindowAttribute(void *hwnd, unsigned long dwAttribute, void *pvAttribute, unsigned long cbAttribute);
static void setWindowsWindowDarkMode(GLFWwindow *window, int value) {
	void *hwnd = glfwGetWin32Window(window);
	// We don't check the return value as this may actually fail on older windows versions,
	// since DWMWA_USE_IMMERSIVE_DARK_MODE is fairly new.
	DwmSetWindowAttribute(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE, &value, sizeof(value));
}
#undef DWMWA_USE_IMMERSIVE_DARK_MODE
#else
static void setWindowsWindowDarkMode(GLFWwindow *window, int iValue) {}
#endif
*/
import "C"
import (
	"fmt"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
)

var onWindowResize func(window *C.GLFWwindow, width int32, height int32)

//export goWindowResizeCallback
func goWindowResizeCallback(window *C.GLFWwindow, width C.int, height C.int) {
	onWindowResize(window, int32(width), int32(height))
}

var onWindowRefresh func(window *C.GLFWwindow)

//export goWindowRefreshCallback
func goWindowRefreshCallback(window *C.GLFWwindow) {
	onWindowRefresh(window)
}

// State contains exported fields, which
// can be changed anytime by the callback
// functions to update certain settings.
type State struct {
	// Target FPS, can be changed anytime seamlessly.
	FrameRate float64
	// GUI Scale and CJK fonts will need a bit of
	// time to load after changing to a different
	// value.
	GUIScale              float32
	LoadCJKFonts          bool
	DarkWindowDecorations bool

	glfwWindow                *C.GLFWwindow
	currGuiScale              float32
	cjkFontsLoaded            bool
	currDarkWindowDecorations bool
}

type Options struct {
	// Initial window size.
	WindowSize imgui.Vec2
	// Minimum window size, or zero.
	WindowMinSize imgui.Vec2
	// Called once, after the initial
	// state is determined and before any
	// drawing takes place. May be nil.
	OnInitWindow func(state *State) error
	// Called to do various pre-frame setup.
	// If anything can err, it should happen
	// in here. No drawing should take place
	// in here. May be nil.
	OnPreDraw func(state *State) error
	// Should not return an error. Control
	// flow should be straight-forward and
	// always reach the end.
	OnDraw func(state *State)
	// Drag-and-drop drop callback, or nil.
	OnDrop func(paths []string)
	// Make the OpenGL context a debug context.
	GLDebugContext bool
}

func Main(title string, options Options) error {
	state := State{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	currentBackend, err := backend.CreateBackend(glfwbackend.NewGLFWBackend())
	if err != nil {
		return fmt.Errorf("creating backend: %w", err)
	}
	defer C.glfwTerminate()

	currentBackend.SetAfterCreateContextHook(func() {
		state.glfwWindow = C.glfwGetCurrentContext()
	})

	if options.GLDebugContext {
		const GLFW_CONTEXT_DEBUG = 0x00022007
		currentBackend.SetWindowFlags(GLFW_CONTEXT_DEBUG, 1)
	}
	currentBackend.SetWindowFlags(glfwbackend.GLFWWindowFlagsResizable, 1)
	currentBackend.CreateWindow(title, int(options.WindowSize.X), int(options.WindowSize.Y))
	if options.WindowMinSize != imgui.NewVec2(0, 0) {
		currentBackend.SetWindowSizeLimits(int(options.WindowMinSize.X), int(options.WindowMinSize.Y), -1, -1)
	}
	defer func() {
		C.ImGui_ImplOpenGL3_Shutdown()
		C.ImGui_ImplGlfw_Shutdown()
		imgui.DestroyContext()
		C.glfwDestroyWindow(state.glfwWindow)
	}()

	C.glfwMakeContextCurrent(state.glfwWindow)
	C.glfwSwapInterval(0)

	currentBackend.SetDropCallback(func(paths []string) {
		if options.OnDrop != nil {
			options.OnDrop(paths)
		}
	})

	io := imgui.CurrentIO()
	flags := io.ConfigFlags()
	flags |= imgui.ConfigFlagsDockingEnable | imgui.ConfigFlagsViewportsEnable
	io.SetConfigFlags(flags)
	io.SetIniFilename("")
	setupFonts(state.LoadCJKFonts)

	{
		_, yScale := currentBackend.ContentScale()
		state.GUIScale = yScale

		monitor := C.glfwGetPrimaryMonitor()
		videoMode := C.glfwGetVideoMode(monitor)
		state.FrameRate = float64(videoMode.refreshRate)
	}
	if options.OnInitWindow != nil {
		if err := options.OnInitWindow(&state); err != nil {
			return err
		}
	}

	if err := gl.Init(); err != nil {
		return fmt.Errorf("initializing OpenGL: %w", err)
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.FrontFace(gl.CW)

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		var typStr string
		if gltype == gl.DEBUG_TYPE_ERROR {
			typStr = "error: "
		}
		log.Printf("GL: %v%v\n", typStr, message)
	}, nil)

	lastDrawTimestamp := time.Now()
	drawAndPresentFrame := func() {
		C.ImGui_ImplGlfw_NewFrame()
		C.ImGui_ImplOpenGL3_NewFrame()
		C.igNewFrame()
		gl.ClearColor(0.2, 0.2, 0.2, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		options.OnDraw(&state)

		imgui.Render()
		C.ImGui_ImplOpenGL3_RenderDrawData(C.igGetDrawData())

		if imgui.CurrentIO().ConfigFlags()&imgui.ConfigFlagsViewportsEnable != 0 {
			prevContext := C.glfwGetCurrentContext()
			imgui.UpdatePlatformWindows()
			imgui.RenderPlatformWindowsDefault()
			C.glfwMakeContextCurrent(prevContext)
		}

		C.glfwSwapBuffers(state.glfwWindow)

		targetFrameTime := time.Duration(float64(time.Second) / state.FrameRate)
		lastDrawTimestamp = lastDrawTimestamp.Add(targetFrameTime)
	}

	C.glfwSetWindowSizeCallback(state.glfwWindow, (*[0]byte)(C.goWindowResizeCallback))
	onWindowResize = func(window *C.GLFWwindow, width, height int32) {
		gl.Viewport(0, 0, width, height)
	}
	C.glfwSetWindowRefreshCallback(state.glfwWindow, (*[0]byte)(C.goWindowRefreshCallback))
	onWindowRefresh = func(window *C.GLFWwindow) {
		drawAndPresentFrame()
	}

	for C.glfwWindowShouldClose(state.glfwWindow) == 0 {
		if options.OnPreDraw != nil {
			if err := options.OnPreDraw(&state); err != nil {
				return err
			}
		}

		if state.GUIScale != state.currGuiScale ||
			state.LoadCJKFonts != state.cjkFontsLoaded ||
			state.DarkWindowDecorations != state.currDarkWindowDecorations {
			if state.LoadCJKFonts != state.cjkFontsLoaded {
				setupFonts(state.LoadCJKFonts)
			}
			io.Ctx().SetStyle(*makeStyle(state.GUIScale))
			if state.DarkWindowDecorations {
				C.setWindowsWindowDarkMode(state.glfwWindow, 1)
			} else {
				C.setWindowsWindowDarkMode(state.glfwWindow, 0)
			}
			state.currGuiScale = state.GUIScale
			state.cjkFontsLoaded = state.LoadCJKFonts
			state.currDarkWindowDecorations = state.DarkWindowDecorations
		}

		timeToDraw := time.Since(lastDrawTimestamp)
		numFramesToDraw := timeToDraw.Seconds() * state.FrameRate
		if timeToDraw < 0 {
			// Frame over-draw
			//log.Printf("Skipped %.5f frames due to over-draw", -numFramesToDraw)
			lastDrawTimestamp = time.Now()
		} else if timeToDraw >= time.Second {
			// Frame under-draw
			//log.Printf("Dropped %.5f frames due to lag", numFramesToDraw)
			lastDrawTimestamp = time.Now()
		} else if numFramesToDraw >= 1 {
			C.glfwPollEvents()
			drawAndPresentFrame()
		} else {
			time.Sleep(timeToDraw)
		}
	}

	return nil
}
