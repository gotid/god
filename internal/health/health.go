package health

import (
	"fmt"
	"github.com/gotid/god/lib/syncx"
	"net/http"
	"strings"
	"sync"
)

// 是一个全局组合式的健康管理器。
var defaultHealthManager = newComboHealthManager()

type (
	// Probe 表示给定组件的就绪状态。
	Probe interface {
		// MarkReady 将端点处理器设置为就绪状态。
		MarkReady()
		// MarkNotReady 将端点处理器设置为未就绪状态。
		MarkNotReady()
		// IsReady 返回组件的内部就绪状态。
		IsReady() bool
		// Name 返回 Probe 的标示名称。
		Name() string
	}

	// healthManager 管理 app 的健康状况。
	healthManager struct {
		ready syncx.AtomicBool
		name  string
	}

	// comboHealthManager 将给定的 probes 组合为一个，以线程安全的方式反映他们的状态。
	comboHealthManager struct {
		mu     sync.Mutex
		probes []Probe
	}
)

// CreateHttpHandler 基于给定的 Probe 创建 http 健康处理程序。
func CreateHttpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if defaultHealthManager.IsReady() {
			_, _ = w.Write([]byte("OK"))
		} else {
			http.Error(w, "服务不可用\n"+defaultHealthManager.verboseInfo(), http.StatusServiceUnavailable)
		}
	}
}

// AddProbe 添加组件状态到全局 comboHealthManager。
func AddProbe(probe Probe) {
	defaultHealthManager.addProbe(probe)
}

func NewHealthManager(name string) Probe {
	return &healthManager{name: name}
}

func (h *healthManager) MarkReady() {
	h.ready.Set(true)
}

func (h *healthManager) MarkNotReady() {
	h.ready.Set(false)
}

func (h *healthManager) IsReady() bool {
	return h.ready.True()
}

func (h *healthManager) Name() string {
	return h.name
}

func newComboHealthManager() *comboHealthManager {
	return &comboHealthManager{}
}

func (m *comboHealthManager) MarkReady() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, probe := range m.probes {
		probe.MarkReady()
	}
}

func (m *comboHealthManager) MarkNotReady() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, probe := range m.probes {
		probe.MarkNotReady()
	}
}

func (m *comboHealthManager) IsReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, probe := range m.probes {
		if !probe.IsReady() {
			return false
		}
	}

	return true
}

func (m *comboHealthManager) verboseInfo() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var info strings.Builder
	for _, probe := range m.probes {
		if probe.IsReady() {
			info.WriteString(fmt.Sprintf("%s 已就绪\n", probe.Name()))
		} else {
			info.WriteString(fmt.Sprintf("%s 未就绪\n", probe.Name()))
		}
	}

	return info.String()
}

// 添加组件状态
func (m *comboHealthManager) addProbe(probe Probe) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.probes = append(m.probes, probe)
}
