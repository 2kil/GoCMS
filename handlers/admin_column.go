package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func buildColumnTree() (map[uint]*models.Column, []models.Column) {
	var all []models.Column
	database.DB.Preload("Parent").Order("sort_order asc").Find(&all)

	columnMap := make(map[uint]*models.Column)
	for i := range all {
		all[i].Children = nil
		columnMap[all[i].ID] = &all[i]
	}

	for i := range all {
		if all[i].ParentID != nil {
			if parent, ok := columnMap[*all[i].ParentID]; ok {
				parent.Children = append(parent.Children, *columnMap[all[i].ID])
			}
		}
	}

	var buildNode func(*models.Column) models.Column
	buildNode = func(column *models.Column) models.Column {
		node := *column
		node.Children = make([]models.Column, 0, len(column.Children))
		for i := range column.Children {
			if child, ok := columnMap[column.Children[i].ID]; ok {
				node.Children = append(node.Children, buildNode(child))
			}
		}
		return node
	}

	var roots []models.Column
	for i := range all {
		if all[i].ParentID == nil {
			roots = append(roots, buildNode(columnMap[all[i].ID]))
		} else if _, ok := columnMap[*all[i].ParentID]; !ok {
			roots = append(roots, buildNode(columnMap[all[i].ID]))
		}
	}
	return columnMap, roots
}

func loadColumnsTree() []models.Column {
	_, roots := buildColumnTree()

	var flatten func(columns []models.Column, depth int, result *[]models.Column)
	flatten = func(columns []models.Column, depth int, result *[]models.Column) {
		for _, c := range columns {
			prefix := ""
			for i := 0; i < depth; i++ {
				prefix += "— "
			}
			c.Name = prefix + c.Name
			*result = append(*result, c)
			flatten(c.Children, depth+1, result)
		}
	}

	var result []models.Column
	flatten(roots, 0, &result)
	return result
}

func loadColumnsTreeFull() []models.Column {
	_, roots := buildColumnTree()
	return roots
}

func ListColumns(c *gin.Context) {
	columns := loadColumnsTreeFull()

	c.HTML(http.StatusOK, "columns.html", gin.H{
		"columns":  columns,
		"nickname": c.MustGet("nickname"),
	})
}

func ShowColumnEdit(c *gin.Context) {
	idStr := c.Param("id")
	var column models.Column

	if idStr != "" && idStr != "new" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || database.DB.First(&column, id).Error != nil {
			c.Redirect(http.StatusFound, "/adm1n/columns")
			return
		}
	}

	flatColumns := loadColumnsTree()
	if column.ID != 0 {
		flatColumns = filterColumnForParentFlat(flatColumns, column.ID)
	}

	var parentID uint
	if column.ParentID != nil {
		parentID = *column.ParentID
	}

	c.HTML(http.StatusOK, "column_edit.html", gin.H{
		"column":           column,
		"all_columns":      flatColumns,
		"parent_id":        parentID,
		"page_templates":   frontTemplatePageFiles(frontCache.getSettings()),
		"current_template": selectedFrontTemplate(frontCache.getSettings()),
		"nickname":         c.MustGet("nickname"),
		"error":            c.Query("error"),
	})
}

func filterColumnForParent(columns []models.Column, excludeID uint) []models.Column {
	var result []models.Column
	for _, c := range columns {
		if c.ID != excludeID {
			children := filterColumnForParent(c.Children, excludeID)
			c.Children = children
			result = append(result, c)
		}
	}
	return result
}

func filterColumnForParentFlat(columns []models.Column, excludeID uint) []models.Column {
	var result []models.Column
	for _, c := range columns {
		if c.ID != excludeID && !isDescendantColumn(c.ID, excludeID) {
			result = append(result, c)
		}
	}
	return result
}

func renderColumnEditWithError(c *gin.Context, column models.Column, message string) {
	flatColumns := loadColumnsTree()
	if column.ID != 0 {
		flatColumns = filterColumnForParentFlat(flatColumns, column.ID)
	}
	var parentID uint
	if column.ParentID != nil {
		parentID = *column.ParentID
	}
	c.HTML(http.StatusOK, "column_edit.html", gin.H{
		"column":           column,
		"all_columns":      flatColumns,
		"parent_id":        parentID,
		"page_templates":   frontTemplatePageFiles(frontCache.getSettings()),
		"current_template": selectedFrontTemplate(frontCache.getSettings()),
		"nickname":         c.MustGet("nickname"),
		"error":            message,
	})
}

func isDescendantColumn(candidateID uint, columnID uint) bool {
	if candidateID == 0 || columnID == 0 {
		return false
	}

	currentID := candidateID
	for currentID != 0 {
		if currentID == columnID {
			return true
		}

		var current models.Column
		if err := database.DB.First(&current, currentID).Error; err != nil || current.ParentID == nil {
			return false
		}
		currentID = *current.ParentID
	}

	return false
}

func SaveColumn(c *gin.Context) {
	idStr := c.Param("id")
	name := c.PostForm("name")
	slug := c.PostForm("slug")
	content := c.PostForm("content")
	pageTemplate := c.PostForm("page_template")
	isPage := c.PostForm("is_page") == "on"
	sortOrderStr := c.PostForm("sort_order")
	parentIDStr := c.PostForm("parent_id")

	var column models.Column
	if idStr != "" && idStr != "new" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || database.DB.First(&column, id).Error != nil {
			c.Redirect(http.StatusFound, "/adm1n/columns")
			return
		}
	}

	column.Name = name
	if slug == "" {
		slug = uuid.New().String()
	}
	column.Slug = slug
	column.IsPage = isPage
	if isPage {
		if pageTemplate != "" {
			column.PageTemplate = pageTemplate
		} else {
			column.PageTemplate = ""
		}
	} else {
		column.PageTemplate = ""
	}
	column.Content = content
	sortOrder, _ := strconv.Atoi(sortOrderStr)
	if sortOrder == 0 && column.ID == 0 {
		var maxSort int
		database.DB.Model(&models.Column{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSort)
		sortOrder = maxSort + 1
	}
	column.SortOrder = sortOrder

	if parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 64)
		if err == nil && (isDescendantColumn(uint(pid), column.ID) || uint(pid) == column.ID) {
			renderColumnEditWithError(c, column, "父栏目不能选择当前栏目或其子栏目")
			return
		}
		if err == nil {
			pidUint := uint(pid)
			column.ParentID = &pidUint
		}
	} else {
		column.ParentID = nil
	}

	var err error
	if column.ID != 0 {
		err = database.DB.Save(&column).Error
	} else {
		err = database.DB.Create(&column).Error
	}
	if err != nil {
		log.Printf("保存栏目失败: %v", err)
		renderColumnEditWithError(c, column, "保存失败: "+err.Error())
		return
	}

	InvalidateCache()
	GenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/columns")
}

func DeleteColumn(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/columns")
		return
	}
	var column models.Column
	if database.DB.First(&column, id).Error != nil {
		c.Redirect(http.StatusFound, "/adm1n/columns")
		return
	}
	database.DB.Model(&models.Column{}).Where("parent_id = ?", id).Update("parent_id", nil)
	database.DB.Delete(&column)
	InvalidateCache()
	GenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/columns")
}

type columnReorderRequest struct {
	IDs []uint `json:"ids"`
}

func ReorderColumns(c *gin.Context) {
	var req columnReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排序数据"})
		return
	}

	columnMap, roots := buildColumnTree()
	orderedRoots := make([]models.Column, 0, len(roots))
	used := make(map[uint]bool, len(roots))
	for _, id := range req.IDs {
		if root, ok := columnMap[id]; ok && root.ParentID == nil && !used[id] {
			orderedRoots = append(orderedRoots, *root)
			used[id] = true
		}
	}
	for _, root := range roots {
		if !used[root.ID] {
			orderedRoots = append(orderedRoots, root)
		}
	}

	for idx, column := range orderedRoots {
		if err := database.DB.Model(&models.Column{}).Where("id = ?", column.ID).Update("sort_order", idx+1).Error; err != nil {
			log.Printf("更新栏目排序失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存排序失败"})
			return
		}
	}

	InvalidateCache()
	GenerateStatic()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
