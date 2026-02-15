package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"
	"time"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB

	// TargetVolumeLarge Цель: 1 ГБ для больших файлов
	TargetVolumeLarge = 1 * GB

	// IterationsSmall 100 повторений для маленьких
	IterationsSmall = 100

	// SmallFileLimit Граница "маленького" файла
	SmallFileLimit = 1 * MB
)

type TestSubject struct {
	Name string
	Path string
}

// ПУТИ (относительно папки bench/)
var programs = []TestSubject{
	{Name: "cpC", Path: "../cpc/cpc.exe"},
	{Name: "cpW", Path: "../cpw/cpw.exe"},
	{Name: "cpCF", Path: "../cpcf/cpcf.exe"},
}

var fileSizes = []struct {
	name string
	size int64
}{
	{"10KB", 10 * KB},
	{"100KB", 100 * KB},
	{"1MB", 1 * MB},
	{"10MB", 10 * MB},
	{"100MB", 100 * MB},
}

func main() {
	for i, prog := range programs {
		absPath, err := filepath.Abs(prog.Path)
		if err != nil {
			fmt.Println("Error path:", err)
			return
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Printf("ERROR: Program not found: %s\nPath: %s\n", prog.Name, absPath)
			return
		}
		programs[i].Path = absPath
	}

	fmt.Println("=======================================================================")
	fmt.Println("BENCHMARK STARTED")
	fmt.Printf("Small files (<= 1MB): Fixed %d iterations\n", IterationsSmall)
	fmt.Printf("Large files (> 1MB):  Target volume ~%d MB\n", TargetVolumeLarge/MB)
	fmt.Println("=======================================================================")
	fmt.Println()

	// Настраиваем TabWriter:
	// minwidth=10 (минимальная ширина колонки)
	// tabwidth=0
	// padding=3 (отступ между колонками)
	// padchar=' ' (заполнитель пробел)
	// flags=0 (без отладочных линий)
	w := tabwriter.NewWriter(os.Stdout, 10, 0, 3, ' ', 0)

	fmt.Fprintln(w, "FILE SIZE\tPROGRAM\tITERS\tDATA VOL\tTIME TOTAL\tSPEED")
	fmt.Fprintln(w, "---------\t-------\t-----\t--------\t----------\t-----")

	for _, fs := range fileSizes {
		var iterations int
		if fs.size <= SmallFileLimit {
			iterations = IterationsSmall
		} else {
			iterations = int(TargetVolumeLarge / fs.size)
			if iterations == 0 {
				iterations = 1
			}
		}

		srcInfo := fmt.Sprintf("bench_src_%s.bin", fs.name)
		dstInfo := fmt.Sprintf("bench_dst_%s.bin", fs.name)

		if err := createDummyFile(srcInfo, fs.size); err != nil {
			fmt.Printf("Error creating source file: %v\n", err)
			continue
		}

		for _, prog := range programs {
			// Небольшая пауза для стабилизации ОС
			time.Sleep(100 * time.Millisecond)

			start := time.Now()

			for i := 0; i < iterations; i++ {
				cmd := exec.Command(prog.Path, srcInfo, dstInfo)
				if err := cmd.Run(); err != nil {
					fmt.Printf("\nError in %s: %v\n", prog.Name, err)
					break
				}
				// Удаляем копию сразу
				os.Remove(dstInfo)
			}

			duration := time.Since(start)

			totalMB := (float64(fs.size) * float64(iterations)) / float64(MB)
			seconds := duration.Seconds()
			speed := 0.0
			if seconds > 0.0001 {
				speed = totalMB / seconds
			}

			fmt.Fprintf(w, "%s\t%s\t%d\t%.1f MB\t%v\t%.2f MB/s\n",
				fs.name,
				prog.Name,
				iterations,
				totalMB,
				duration.Round(time.Millisecond),
				speed,
			)
		}

		// Разделитель между группами размеров (пустая строка для читаемости)
		fmt.Fprintln(w, "\t\t\t\t\t")

		// Сбрасываем буфер на экран после каждого размера файла
		w.Flush()

		// Удаляем исходник
		os.Remove(srcInfo)
	}

	fmt.Println("Done.")
}

func createDummyFile(filename string, size int64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	bufSize := 1 * MB
	if size < int64(bufSize) {
		bufSize = int(size)
	}

	buf := make([]byte, bufSize)
	// Используем math/rand, он быстрее crypto/rand для заполнения мусором
	rand.Read(buf)

	written := int64(0)
	for written < size {
		toWrite := int64(len(buf))
		if size-written < toWrite {
			toWrite = size - written
		}
		if _, err := f.Write(buf[:toWrite]); err != nil {
			return err
		}
		written += toWrite
	}
	return nil
}
