package filemanager

import "os"

type FileManager struct {
	file *os.File
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (fm *FileManager) ReadAt(offset int64, buffer []byte) error {
	n, err := fm.file.ReadAt(buffer, offset)

	if n != len(buffer) {
		return ErrShortRead
	}
	return err
}

func (fm *FileManager) WriteAt(offset int64, buffer []byte) error {
	n, err := fm.file.WriteAt(buffer, offset)

	if n != len(buffer) {
		return ErrShortWrite
	}
	return err
}

func (fm *FileManager) Sync() error {
	return fm.file.Sync()
}

func (fm *FileManager) Size() (int64, error) {
	stat, err := fm.file.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func Open(path string) (*FileManager, error) {
	file, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		0644,
	)

	if err != nil {
		return nil, err
	}

	return &FileManager{
		file: file,
	}, nil
}

func (fm *FileManager) Close() error {
	return fm.file.Close()
}

func (fm *FileManager) Truncate(size int64) error {
	return fm.file.Truncate(size)
}
