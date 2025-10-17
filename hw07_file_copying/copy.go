package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	//32 KB
	bufferSize := 32 * 1024

	file, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	fileState, err := file.Stat()
	if err != nil {
		return err
	}
	if !fileState.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	if offset < 0 {
		return errors.New("negative offset")
	}
	if limit < 0 {
		return errors.New("negative limit")
	}

	fileSize := fileState.Size()
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	data := make([]byte, bufferSize)
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	outFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer func(out *os.File) { _ = out.Close() }(outFile)

	var remaining int64
	if limit == 0 {
		remaining = fileSize - offset
	} else {
		remaining = limit
		if remaining > fileSize-offset {
			remaining = fileSize - offset
		}
	}

	totalToCopy := remaining
	lastPct := -1
	printProgress := func(done, total int64) {
		if total <= 0 {
			return
		}
		pct := int(done * 100 / total)
		if pct != lastPct {
			lastPct = pct
			fmt.Printf("\rProgress: %d%%", pct)
		}
	}

	for remaining > 0 {
		toRead := data
		if int64(len(toRead)) > remaining {
			toRead = data[:remaining]
		}

		n, err := file.Read(toRead)
		if n > 0 {
			if _, werr := outFile.Write(toRead[:n]); werr != nil {
				return werr
			}
			remaining -= int64(n)
			if totalToCopy > 0 {
				done := totalToCopy - remaining
				printProgress(done, totalToCopy)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if totalToCopy > 0 {
		fmt.Printf("\rProgress: %3d%%\n", 100)
	}
	return nil
}
