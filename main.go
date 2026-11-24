package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof" // Автоматическая регистрация роутов /debug/pprof/
	"runtime"
	"time"
)

// 1. Класс генерации нагрузки на память (Heap)
type MemoryAllocatorService struct {
	blocks [][]byte
}

func (m *MemoryAllocatorService) RunAllocationLoop(preventLeak bool) {
	for {
		// Выделяем блок 1 МБ и заполняем его, чтобы ОС реально выделила страницы памяти
		block := make([]byte, 1024*1024)
		block[0] = 1
		m.blocks = append(m.blocks, block)

		// Исправление утечки (Шаг 5): сбрасываем ссылки каждые 10 МБ
		if preventLeak && len(m.blocks) > 10 {
			m.blocks = nil
			runtime.GC()
		}
		time.Sleep(60 * time.Millisecond)
	}
}

// 2. Класс генерации высокой нагрузки на процессор (Обязательное задание №1)
type CPUIntensiveCalculator struct {
	Acc float64
}

func (c *CPUIntensiveCalculator) RunHeavyMathLoop() {
	for {
		// Тяжёлые тригонометрические вычисления в цикле
		val := 0.0
		for i := 1; i < 500_000; i++ {
			val += math.Sin(float64(i)) * math.Cos(float64(i)) / math.Sqrt(float64(i))
		}
		c.Acc = val
		time.Sleep(5 * time.Millisecond) // Короткая пауза для кооперации с планировщиком
	}
}

// 3. Класс мониторинга текущего состояния памяти в терминале
type TelemetryReporter struct{}

func (TelemetryReporter) StartConsoleLog() {
	var st runtime.MemStats
	for range time.Tick(1 * time.Second) {
		runtime.ReadMemStats(&st)
		fmt.Printf("\r📊 В куче (HeapAlloc): %4d МБ | Горутин: %d   ",
			st.HeapAlloc/(1024*1024), runtime.NumGoroutine())
	}
}

func main() {
	fixFlag := flag.Bool("fix", false, "Включить очистку среза (устранить утечку памяти)")
	port := flag.String("port", ":6060", "Порт сервера pprof")
	flag.Parse()

	memService := &MemoryAllocatorService{}
	cpuService := &CPUIntensiveCalculator{}
	reporter := TelemetryReporter{}

	go memService.RunAllocationLoop(*fixFlag)
	go cpuService.RunHeavyMathLoop()
	go reporter.StartConsoleLog()

	mode := "❌ РЕЖИМ УТЕЧКИ ПАМЯТИ"
	if *fixFlag {
		mode = "✅ РЕЖИМ БЕЗ УТЕЧКИ (ИСПРАВЛЕНО)"
	}

	fmt.Printf("🚀 Профайлер pr18 запущен (%s)\n", mode)
	fmt.Printf("🌐 Веб-интерфейс: http://localhost%s/debug/pprof/\n\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}
