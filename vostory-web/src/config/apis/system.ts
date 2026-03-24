import request from "@/packages/request";

export type PostListDetailType = {
    id: number;
    post_code: string;
    post_name: string;
    post_sort: number;
    status: string;
    remark: string;
    created_at: string;
    updated_at: string;
};

export type PostListParams = {
    page?: number;
    size?: number;
    post_code?: string;
    post_name?: string;
    status?: string;
};

export type PostCreateParams = {
    post_code: string;
    post_name: string;
    post_sort: number;
    status: string;
    remark?: string;
};

export type PostUpdateParams = PostCreateParams & {
    id: number;
};

/**
 * 获取岗位列表
 * @param params
 * @returns
 */
export function getPostList(params?: PostListParams): Promise<{
    data: PostListDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({
        url: "/api/v1/system/post/list",
        params
    });
}

/**
 * 获取岗位详情
 * @param id
 * @returns
 */
export function getPost(id: number): Promise<PostListDetailType> {
    return request({
        url: `/api/v1/system/post/${id}`
    });
}

/**
 * 添加岗位
 * @param data
 * @returns
 */
export function addPost(data: PostCreateParams) {
    return request({
        url: "/api/v1/system/post",
        method: "post",
        data
    });
}

/**
 * 更新岗位
 * @param data
 * @returns
 */
export function updatePost(data: PostUpdateParams) {
    return request({
        url: `/api/v1/system/post/${data.id}`,
        method: "put",
        data
    });
}

/**
 * 删除岗位
 * @param id
 * @returns
 */
export function deletePost(id: number) {
    return request({
        url: `/api/v1/system/post/${id}`,
        method: "delete"
    });
}

export function enablePost(id: number) {
    return request({
        url: `/api/v1/system/post/${id}/enable`,
        method: "put"
    });
}

export function disablePost(id: number) {
    return request({
        url: `/api/v1/system/post/${id}/disable`,
        method: "put"
    });
}

// ============= 部门相关接口 =============

// 用户简要信息类型
export type UserBriefType = {
    user_id: number;
    user_name: string;
    avatar: string;
};

export type DeptDetailType = {
    id: number;
    parent_id: number;
    ancestors: string;
    dept_name: string;
    order_num: number;
    leader_id?: number;
    leader: string;
    leader_user?: UserBriefType;
    phone: string;
    email: string;
    status: string;
    remark: string;
    created_at: string;
    updated_at: string;
    children?: DeptDetailType[];
};

export type DeptTreeType = {
    id: number;
    parent_id: number;
    dept_name: string;
    order_num: number;
    leader_id?: number;
    leader: string;
    leader_user?: UserBriefType;
    status: string;
    children?: DeptTreeType[];
};

export type DeptListParams = {
    current?: number;
    size?: number;
    dept_name?: string;
    status?: string;
    parent_id?: number;
};

export type DeptCreateParams = {
    parent_id: number;
    dept_name: string;
    order_num: number;
    leader_id?: number;
    leader?: string;
    phone?: string;
    email?: string;
    status: string;
    remark?: string;
};

export type DeptUpdateParams = DeptCreateParams & {
    id: number;
};

/**
 * 获取部门列表（分页）
 * @param params
 * @returns
 */
export function getDeptList(params?: DeptListParams): Promise<{
    records: DeptDetailType[];
    total: number;
}> {
    return request({
        url: "/api/v1/system/dept/list",
        params
    });
}

/**
 * 获取部门树
 * @returns
 */
export function getDeptTree(params: any): Promise<DeptTreeType[]> {
    return request({
        url: "/api/v1/system/dept/tree",
        params
    });
}

/**
 * 获取部门详情
 * @param id
 * @returns
 */
export function getDept(id: number): Promise<DeptDetailType> {
    return request({
        url: `/api/v1/system/dept/${id}`
    });
}

/**
 * 添加部门
 * @param data
 * @returns
 */
export function addDept(data: DeptCreateParams) {
    return request({
        url: "/api/v1/system/dept",
        method: "post",
        data
    });
}

/**
 * 更新部门
 * @param data
 * @returns
 */
export function updateDept(data: DeptUpdateParams) {
    return request({
        url: `/api/v1/system/dept/${data.id}`,
        method: "put",
        data
    });
}

/**
 * 删除部门
 * @param id
 * @returns
 */
export function deleteDept(id: number) {
    return request({
        url: `/api/v1/system/dept/${id}`,
        method: "delete"
    });
}

export function enableDept(id: number) {
    return request({
        url: `/api/v1/system/dept/${id}/enable`,
        method: "put"
    });
}

export function disableDept(id: number) {
    return request({
        url: `/api/v1/system/dept/${id}/disable`,
        method: "put"
    });
}

// ============= 菜单相关接口 =============

export type MenuDetailType = {
    id: number;
    parent_id: number;
    menu_name: string;
    order_num: number;
    url: string;
    target: string;
    menu_type: string;
    visible: string;
    is_refresh: string;
    perms: string;
    icon: string;
    create_by: string;
    update_by: string;
    remark: string;
    created_at: string;
    updated_at: string;
    children?: MenuDetailType[];
};

export type MenuTreeType = {
    id: number;
    parent_id: number;
    menu_name: string;
    order_num: number;
    url: string;
    menu_type: string;
    visible: string;
    perms: string;
    icon: string;
    method?: string;
    children?: MenuTreeType[];
};

export type MenuListParams = {
    current?: number;
    size?: number;
    menu_name?: string;
    visible?: string;
    parent_id?: number;
};

export type MenuCreateParams = {
    parent_id: number;
    menu_name: string;
    order_num: number;
    url?: string;
    target?: string;
    menu_type: string;
    visible: string;
    is_refresh: string;
    perms?: string;
    icon?: string;
    remark?: string;
};

export type MenuUpdateParams = MenuCreateParams & {
    id: number;
};

/**
 * 获取菜单列表（分页）
 * @param params
 * @returns
 */
export function getMenuList(params?: MenuListParams): Promise<{
    records: MenuDetailType[];
    total: number;
}> {
    return request({
        url: "/api/v1/system/menu/list",
        params
    });
}

/**
 * 获取菜单树
 * @returns
 */
export function getMenuTree(params: any): Promise<MenuTreeType[]> {
    return request({
        url: "/api/v1/system/menu/tree",
        params
    });
}

/**
 * 根据类型获取菜单
 * @param menuType
 * @returns
 */
export function getMenusByType(menuType: string): Promise<MenuDetailType[]> {
    return request({
        url: `/api/v1/system/menu/type/${menuType}`
    });
}

/**
 * 获取菜单详情
 * @param id
 * @returns
 */
export function getMenu(id: number): Promise<MenuDetailType> {
    return request({
        url: `/api/v1/system/menu/${id}`
    });
}

/**
 * 添加菜单
 * @param data
 * @returns
 */
export function addMenu(data: MenuCreateParams) {
    return request({
        url: "/api/v1/system/menu",
        method: "post",
        data
    });
}

/**
 * 更新菜单
 * @param data
 * @returns
 */
export function updateMenu(data: MenuUpdateParams) {
    return request({
        url: `/api/v1/system/menu/${data.id}`,
        method: "put",
        data
    });
}

/**
 * 删除菜单
 * @param id
 * @returns
 */
export function deleteMenu(id: number) {
    return request({
        url: `/api/v1/system/menu/${id}`,
        method: "delete"
    });
}

/**
 * 批量添加权限
 * @param data
 * @returns
 */
export function muAddPers(data: { parent_id: number; menu_name: string; perms: string }[]) {
    return request({
        url: `/api/v1/system/menu/perms/muti`,
        method: "post",
        data
    });
}

// ============= 角色相关接口 =============

export type RoleDetailType = {
    role_id: number;
    role_name: string;
    role_key: string;
    role_sort: number;
    data_scope: string;
    status: string;
    create_by: string;
    update_by: string;
    remark: string;
    created_at: string;
    updated_at: string;
};

export type RoleListParams = {
    current?: number;
    size?: number;
    role_name?: string;
    role_key?: string;
    status?: string;
};

export type RoleCreateParams = {
    role_name: string;
    role_key: string;
    role_sort: number;
    data_scope: string;
    status: string;
    remark?: string;
};

export type RoleUpdateParams = RoleCreateParams & {
    id: number;
};

/**
 * 获取角色列表（分页）
 * @param params
 * @returns
 */
export function getRoleList(params?: RoleListParams): Promise<{
    data: RoleDetailType[];
    total: number;
}> {
    return request({
        url: "/api/v1/system/role/list",
        params
    });
}

/**
 * 获取角色详情
 * @param id
 * @returns
 */
export function getRole(id: number): Promise<RoleDetailType> {
    return request({
        url: `/api/v1/system/role/${id}`
    });
}

/**
 * 添加角色
 * @param data
 * @returns
 */
export function addRole(data: RoleCreateParams) {
    return request({
        url: "/api/v1/system/role",
        method: "post",
        data
    });
}

/**
 * 更新角色
 * @param data
 * @returns
 */
export function updateRole(data: RoleUpdateParams) {
    return request({
        url: `/api/v1/system/role/${data.id}`,
        method: "put",
        data
    });
}

/**
 * 删除角色
 * @param id
 * @returns
 */
export function deleteRole(id: number) {
    return request({
        url: `/api/v1/system/role/${id}`,
        method: "delete"
    });
}

export function enableRole(id: number) {
    return request({
        url: `/api/v1/system/role/${id}/enable`,
        method: "put"
    });
}

export function disableRole(id: number) {
    return request({
        url: `/api/v1/system/role/${id}/disable`,
        method: "put"
    });
}

export function getRoleMenuDetail(id: string | number): Promise<number[]> {
    return request({
        url: `/api/v1/system/role/${id}/menus`
    });
}

export function updateRoleMenu(data: { menu_ids: number[]; role_id: number }) {
    return request({
        url: `/api/v1/system/role/${data.role_id}/menus`,
        method: "put",
        data
    });
}

export type RoleListDetailType = {
    roleId: string;
    roleName: string;
    roleCode: string;
    roleDesc: string;
    dsType: number;
    dsScope: string;
    createBy: string;
    updateBy: string;
    createTime: string;
    updateTime: string;
    delFlag: string;
};

// ============= 用户相关接口 =============

export type UserDetailType = {
    user_id: number;
    dept_id?: number;
    dept_name?: string;
    superior_id?: number;
    superior_name?: string;
    login_name: string;
    user_name: string;
    user_type: string;
    email: string;
    phonenumber: string;
    sex: string;
    avatar: string;
    status: string;
    login_ip: string;
    login_date?: string;
    pwd_update_date?: string;
    create_by: string;
    update_by: string;
    remark: string;
    created_at: string;
    updated_at: string;
    role_ids?: number[];
    post_ids?: number[];
    roles?: RoleListDetailType[];
    posts?: PostListDetailType[];
    superior?: {
        user_id: number;
        user_name: string;
    };
};

export type UserListParams = {
    current?: number;
    size?: number;
    login_name?: string;
    user_name?: string;
    status?: string;
    dept_id?: number;
    phonenumber?: string;
    email?: string;
};

export type UserCreateParams = {
    dept_id?: number;
    superior_id?: number;
    login_name: string;
    user_name: string;
    user_type?: string;
    email?: string;
    phonenumber?: string;
    sex?: string;
    avatar?: string;
    password: string;
    status: string;
    role_ids?: number[];
    post_ids?: number[];
    remark?: string;
};

export type UserUpdateParams = Omit<UserCreateParams, "password"> & {
    user_id: number;
    password?: string; // 更新时密码可选
};

export type UserResetPasswordParams = {
    new_password: string;
};

export type UserChangeStatusParams = {
    status: string;
};

/**
 * 获取用户列表（分页）
 * @param params
 * @returns
 */
export function getUserList(params?: UserListParams): Promise<{
    data: UserDetailType[];
    total: number;
    page: number;
    size: number;
}> {
    return request({
        url: "/api/v1/system/user/list",
        params
    });
}

/**
 * 获取用户详情
 * @param id
 * @returns
 */
export function getUser(id: number): Promise<UserDetailType> {
    return request({
        url: `/api/v1/system/user/${id}`
    });
}

/**
 * 添加用户
 * @param data
 * @returns
 */
export function addUser(data: UserCreateParams) {
    return request({
        url: "/api/v1/system/user",
        method: "post",
        data
    });
}

/**
 * 更新用户
 * @param data
 * @returns
 */
export function updateUser(data: UserUpdateParams) {
    return request({
        url: `/api/v1/system/user/${data.user_id}`,
        method: "put",
        data
    });
}

/**
 * 删除用户
 * @param id
 * @returns
 */
export function deleteUser(id: number) {
    return request({
        url: `/api/v1/system/user/${id}`,
        method: "delete"
    });
}

/**
 * 重置用户密码（管理员操作）
 * @param id
 * @param password
 * @returns
 */
export function resetUserPassword(id: number, password: string) {
    return request({
        url: `/api/v1/system/user/${id}/update-password`,
        method: "put",
        data: {
            password
        }
    });
}

/**
 * 当前用户修改密码
 * @param oldPassword - 原密码（加密后）
 * @param newPassword - 新密码（加密后）
 * @returns
 */
export function updateCurrentUserPassword(oldPassword: string, newPassword: string) {
    return request({
        url: `/api/v1/user/update-password`,
        method: "put",
        data: {
            old_password: oldPassword,
            new_password: newPassword
        }
    });
}

/**
 * 启用用户
 * @param id
 * @param data
 * @returns
 */
export function enableUser(id: number) {
    return request({
        url: `/api/v1/system/user/${id}/enable`,
        method: "put"
    });
}

/**
 * 禁用用户
 * @param id
 * @returns
 */
export function disableUser(id: number) {
    return request({
        url: `/api/v1/system/user/${id}/disable`,
        method: "put"
    });
}

// 用户导入错误类型
export type UserImportError = {
    lineNum: number;
    errors: string[];
};

// 用户导入结果类型
export type UserImportResult = {
    success_count: number;
    fail_count: number;
    errors: UserImportError[];
};

/**
 * 导入用户
 * @param file Excel文件
 * @returns
 */
export function importUsers(file: File): Promise<UserImportResult> {
    const formData = new FormData();
    formData.append("file", file);
    return request({
        url: "/api/v1/system/user/import",
        method: "post",
        data: formData,
        headers: {
            "Content-Type": "multipart/form-data"
        }
    });
}

/**
 * 下载用户导入模板
 * @returns
 */
export function downloadUserImportTemplate() {
    return request({
        url: "/api/v1/system/user/import/template",
        method: "get",
        responseType: "blob"
    });
}

/**
 * 获取登录日志列表
 * @param params
 * @returns
 */
export function getLoginInfoList(params: any) {
    return request({
        url: "/api/v1/system/logininfor/list",
        params
    });
}

/**
 * 接口列表
 * @param params
 * @returns
 */
export function getApiList(params: any) {
    return request({
        url: "/api/v1/system/api/list",
        params
    });
}

/**
 * 获取接口标签列表
 * @returns
 */
export function getApiTagList() {
    return request({
        url: "/api/v1/system/api/tag/list"
    });
}

// ============= 通用用户选项接口 =============

/**
 * 用户选项类型（简化版，用于下拉选择）
 */
export type UserOptionType = {
    user_id: number;
    user_name: string;
    login_name: string;
    dept_name?: string;
};

/**
 * 用户选项查询参数
 */
export type UserOptionParams = {
    keyword?: string; // 搜索关键词（用户名/登录名）
    dept_id?: number; // 部门ID
    limit?: number; // 返回数量限制，默认50
};

/**
 * 获取用户选项列表（用于下拉选择，如转办、指定审批人等场景）
 * 此接口只需登录即可访问，无需特定权限
 * @param params
 * @returns
 */
export function getUserOptions(params?: UserOptionParams): Promise<UserOptionType[]> {
    return request({
        url: "/api/v1/common/user/options",
        params
    });
}

// ============= 角色选项接口（白名单） =============

/**
 * 角色选项类型
 */
export type RoleOptionType = {
    role_id: number;
    role_name: string;
};

/**
 * 角色选项查询参数
 */
export type RoleOptionParams = {
    keyword?: string; // 搜索关键词（角色名称）
    limit?: number; // 返回数量限制，默认100
};

/**
 * 获取角色选项列表（用于下拉选择，如指定角色等场景）
 * 此接口只需登录即可访问，无需特定权限
 * @param params
 * @returns
 */
export function getRoleOptions(params?: RoleOptionParams): Promise<RoleOptionType[]> {
    return request({
        url: "/api/v1/common/role/options",
        params
    });
}

// ============= 部门选项接口（白名单） =============

/**
 * 部门选项树类型
 */
export type DeptOptionTreeType = {
    dept_id: number;
    dept_name: string;
    children?: DeptOptionTreeType[];
};

/**
 * 获取部门选项树（用于下拉选择，如指定部门等场景）
 * 此接口只需登录即可访问，无需特定权限
 * @returns
 */
export function getDeptOptions(): Promise<DeptOptionTreeType[]> {
    return request({
        url: "/api/v1/common/dept/options"
    });
}
