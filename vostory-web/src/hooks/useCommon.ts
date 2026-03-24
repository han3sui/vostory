import {
    getUserOptions,
    type UserOptionType,
    getRoleOptions,
    type RoleOptionType,
    getDeptOptions,
    type DeptOptionTreeType
} from "@/config/apis/system";

export const useUserListOptions = () => {
    const userListOptions = ref<{ label: string; value: number }[]>([]);
    const userList = ref<UserOptionType[]>([]);
    const loading = ref(false);

    const getData = async (keyword?: string) => {
        loading.value = true;
        try {
            const res = await getUserOptions({ keyword, limit: 100 });
            userList.value = res;
            userListOptions.value = res.map((u) => ({
                label: `${u.user_name}(${u.login_name})`,
                value: u.user_id
            }));
        } finally {
            loading.value = false;
        }
    };

    onMounted(() => {
        getData();
    });

    return { userListOptions, userList, getData, loading };
};

export const useRoleListOptions = () => {
    const roleListOptions = ref<{ label: string; value: number }[]>([]);
    const roleList = ref<RoleOptionType[]>([]);
    const loading = ref(false);

    const getData = async (keyword?: string) => {
        loading.value = true;
        try {
            const res = await getRoleOptions({ keyword, limit: 500 });
            roleList.value = res || [];
            roleListOptions.value = (res || []).map((r: RoleOptionType) => ({
                label: r.role_name,
                value: r.role_id
            }));
        } finally {
            loading.value = false;
        }
    };

    onMounted(() => {
        getData();
    });

    return { roleListOptions, roleList, getData, loading };
};

// 递归转换部门树为 TreeSelect 需要的格式
const convertDeptTreeToOptions = (tree: DeptOptionTreeType[]): any[] => {
    return tree.map((item) => ({
        key: item.dept_id,
        title: item.dept_name,
        value: item.dept_id,
        children: item.children ? convertDeptTreeToOptions(item.children) : undefined
    }));
};

export const useDeptTreeOptions = () => {
    const deptTreeOptions = ref<any[]>([]);
    const deptTree = ref<DeptOptionTreeType[]>([]);
    const loading = ref(false);

    const getData = async () => {
        loading.value = true;
        try {
            const res = await getDeptOptions();
            deptTree.value = res || [];
            deptTreeOptions.value = convertDeptTreeToOptions(res || []);
        } finally {
            loading.value = false;
        }
    };

    onMounted(() => {
        getData();
    });

    return { deptTreeOptions, deptTree, getData, loading };
};
