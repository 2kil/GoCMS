package handlers

import (
	"testing"

	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func TestUploadReferenceHelpers(t *testing.T) {
	setupSchedulerTestDB(t)

	oldURL := "/upload/202607/a_b%25.jpg"
	newURL := "/upload/202607/renamed.jpg"
	post := models.Post{
		Title:      "引用图片",
		Slug:       "image-ref-post",
		CoverImage: oldURL,
		Content:    `<p><img src="` + oldURL + `" style="width:50%"></p>`,
	}
	otherPost := models.Post{
		Title:   "相似图片",
		Slug:    "similar-image-post",
		Content: `<p><img src="/upload/202607/axb%25.jpg"></p>`,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("创建图片引用文章失败: %v", err)
	}
	if err := database.DB.Create(&otherPost).Error; err != nil {
		t.Fatalf("创建相似图片文章失败: %v", err)
	}

	count, err := CountUploadReferences(oldURL)
	if err != nil {
		t.Fatalf("统计图片引用失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("期望统计 1 篇引用文章，实际为 %d", count)
	}

	if err := ReplaceUploadReferences(oldURL, newURL); err != nil {
		t.Fatalf("替换图片引用失败: %v", err)
	}

	var got models.Post
	if err := database.DB.First(&got, post.ID).Error; err != nil {
		t.Fatalf("读取替换后的文章失败: %v", err)
	}
	if got.CoverImage != newURL {
		t.Fatalf("封面图未替换，得到 %q", got.CoverImage)
	}
	if got.Content != `<p><img src="`+newURL+`" style="width:50%"></p>` {
		t.Fatalf("正文图片未替换，得到 %q", got.Content)
	}

	var gotOther models.Post
	if err := database.DB.First(&gotOther, otherPost.ID).Error; err != nil {
		t.Fatalf("读取相似图片文章失败: %v", err)
	}
	if gotOther.Content != otherPost.Content {
		t.Fatalf("相似图片文章被错误替换，得到 %q", gotOther.Content)
	}
}
