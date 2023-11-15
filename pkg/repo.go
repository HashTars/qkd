package pkg

import "time"

type FileStore struct {
	Id         int64
	UUID       string
	Bucket     string
	RelPath    string
	FileName   string
	Ext        string
	CreateTime time.Time
}

// 插入单个数据
func InsertFileStore(fileStore FileStore) error {
	_, err := MinioHelperIns.db.Exec("INSERT INTO file_store (uuid, bucket,rel_path, file_name,ext) VALUES ($1, $2, $3, $4, $5)",
		fileStore.UUID, fileStore.Bucket, fileStore.RelPath, fileStore.FileName, fileStore.Ext)
	return err
}

// 根据 UUID 获取单个数据
func GetFileStoreByUUID(uuid string) (FileStore, error) {
	var fs FileStore
	err := MinioHelperIns.db.QueryRow("SELECT id, uuid, bucket, rel_path, file_name, ext FROM file_store WHERE uuid = $1", uuid).
		Scan(&fs.Id, &fs.UUID, &fs.Bucket, &fs.RelPath, &fs.FileName, &fs.Ext)
	if err != nil {
		return FileStore{}, err
	}
	return fs, nil
}
