package models

type File struct {
	Id          int64  `json:"id" gorm:"primaryKey;unique;notnull"`
	Cid         string `json:"cid" gorm:"unique;notnull"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	CreatedTime uint64 `json:"created_time" gorm:"autoCreateTime;notnull comment:user创建时间"`
}

func (f *File) Add(file File) (int64, int64, error) {
	result := DB.Create(&file)
	return result.RowsAffected, file.Id, result.Error
}

func (f *File) Count(cid string) bool {
	var count int64
	DB.Model(&File{}).Where("cid = ?", cid).Count(&count)
	if count > 0 {
		return true
	}
	return false
}
