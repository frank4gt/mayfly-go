package api

import (
	"mayfly-go/internal/sys/api/form"
	"mayfly-go/internal/sys/api/vo"
	"mayfly-go/internal/sys/application"
	"mayfly-go/internal/sys/domain/entity"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/contextx"
	"mayfly-go/pkg/ginx"
	"mayfly-go/pkg/req"
	"mayfly-go/pkg/utils/collx"
	"strconv"
	"strings"
)

type Role struct {
	RoleApp     application.Role
	ResourceApp application.Resource
}

func (r *Role) Roles(rc *req.Ctx) {
	g := rc.GinCtx
	condition := &entity.Role{Name: g.Query("name")}
	rc.ResData = r.RoleApp.GetPageList(condition, ginx.GetPageParam(g), new([]entity.Role))
}

// 保存角色信息
func (r *Role) SaveRole(rc *req.Ctx) {
	form := &form.RoleForm{}
	role := ginx.BindJsonAndCopyTo(rc.GinCtx, form, new(entity.Role))
	rc.ReqParam = form
	role.SetBaseInfo(rc.LoginAccount)

	r.RoleApp.SaveRole(role)
}

// 删除角色及其资源关联关系
func (r *Role) DelRole(rc *req.Ctx) {
	idsStr := ginx.PathParam(rc.GinCtx, "id")
	rc.ReqParam = idsStr
	ids := strings.Split(idsStr, ",")

	for _, v := range ids {
		value, err := strconv.Atoi(v)
		biz.ErrIsNilAppendErr(err, "string类型转换为int异常: %s")
		r.RoleApp.DeleteRole(uint64(value))
	}
}

// 获取角色关联的资源id数组，用于分配资源时回显已拥有的资源
func (r *Role) RoleResourceIds(rc *req.Ctx) {
	rc.ResData = r.RoleApp.GetRoleResourceIds(uint64(ginx.PathParamInt(rc.GinCtx, "id")))
}

// 查看角色关联的资源树信息
func (r *Role) RoleResource(rc *req.Ctx) {
	g := rc.GinCtx

	var resources vo.ResourceManageVOList
	r.RoleApp.GetRoleResources(uint64(ginx.PathParamInt(g, "id")), &resources)

	rc.ResData = resources.ToTrees(0)
}

// 保存角色资源
func (r *Role) SaveResource(rc *req.Ctx) {
	var form form.RoleResourceForm
	ginx.BindJsonAndValid(rc.GinCtx, &form)
	rc.ReqParam = form

	// 将,拼接的字符串进行切割并转换
	newIds := collx.ArrayMap[string, uint64](strings.Split(form.ResourceIds, ","), func(val string) uint64 {
		id, _ := strconv.Atoi(val)
		return uint64(id)
	})

	r.RoleApp.SaveRoleResource(contextx.NewLoginAccount(rc.LoginAccount), form.Id, newIds)
}
