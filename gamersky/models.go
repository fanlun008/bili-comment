package gamersky

// NewsInfo 新闻信息结构体
type NewsInfo struct {
	SID         string `json:"sid"`          // 新闻ID (data-sid)
	Title       string `json:"title"`        // 新闻标题
	Time        string `json:"time"`         // 发布时间
	CommentNum  int    `json:"comment_num"`  // 评论数
	URL         string `json:"url"`          // 新闻链接
	ImageURL    string `json:"image_url"`    // 图片链接
	CreateTime  string `json:"create_time"`  // 记录创建时间
	TopLineTime string `json:"topline_time"` // 置顶时间
}

// Comment Gamersky评论信息结构体
type Comment struct {
	ID                 int64  `json:"id"`                   // 评论ID
	ArticleID          string `json:"article_id"`           // 文章ID
	UserID             int    `json:"user_id"`              // 用户ID
	Username           string `json:"username"`             // 用户名
	Content            string `json:"content"`              // 评论内容
	CommentTime        string `json:"comment_time"`         // 评论时间
	SupportCount       int    `json:"support_count"`        // 点赞数
	ReplyCount         int    `json:"reply_count"`          // 回复数
	ParentID           int64  `json:"parent_id"`            // 父评论ID (0表示一级评论)
	AnswerToID         int64  `json:"answer_to_id"`         // 被回复的评论ID (0表示一级评论)
	AnswerToName       string `json:"answer_to_name"`       // 被回复用户名
	UserAvatar         string `json:"user_avatar"`          // 用户头像
	UserLevel          int    `json:"user_level"`           // 用户等级
	IPLocation         string `json:"ip_location"`          // IP位置
	DeviceName         string `json:"device_name"`          // 设备名称
	FloorNumber        int    `json:"floor_number"`         // 楼层号
	IsTuijian          bool   `json:"is_tuijian"`           // 是否推荐
	IsAuthor           bool   `json:"is_author"`            // 是否作者
	IsBest             bool   `json:"is_best"`              // 是否最佳
	UserAuthentication string `json:"user_authentication"`  // 用户认证
	UserGroupID        int    `json:"user_group_id"`        // 用户组ID
	ThirdPlatformBound string `json:"third_platform_bound"` // 第三方平台绑定
	CreateTime         string `json:"create_time"`          // 记录创建时间
}
