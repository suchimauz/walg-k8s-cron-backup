package job

import (
	"encoding/json"
	"time"
)

// Example BackupInfo Json
// {
// 	"backup_name": string
// 	"time": time.Time,
// 	"wal_file_name": string,
// 	"start_time": time.Time,
// 	"finish_time": time.Time,
// 	"date_fmt": string,
// 	"hostname": string,
// 	"data_dir": string,
// 	"pg_version": int,
// 	"start_lsn": int64,
// 	"finish_lsn": int64,
// 	"is_permanent": false,
// 	"system_identifier": int64,
// 	"uncompressed_size": int64,
// 	"compressed_size": int64
// }

// Helper struct for parse wal-g backups info json
type BackupInfo struct {
	BackupName       string    `json:"backup_name"`
	Time             time.Time `json:"time"`
	UncompressedSize int64     `json:"uncompressed_size"`
	CompressedSize   int64     `json:"compressed_size"`
}

// Func for parse json to struct
func parseBackupsInfoJson(backupsJson string) ([]*BackupInfo, error) {
	var backupsInfo []*BackupInfo

	// Parse json into new BackupInfo object
	err := json.Unmarshal([]byte(backupsJson), &backupsInfo)
	if err != nil {
		return nil, err
	}

	return backupsInfo, nil
}
