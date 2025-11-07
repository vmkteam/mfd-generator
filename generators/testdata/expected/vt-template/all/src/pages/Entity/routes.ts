/* eslint-disable */
export default [
  /* Category */
  {
    name: "categoryList",
    path: "/categories",
    component: () =>
      import("@/pages/Entity/Category/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "categoryList"]
    }
  },
  {
    name: "categoryEdit",
    path: "/categories/:id/edit",
    component: () =>
      import("@/pages/Entity/Category/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "categoryList", "categoryEdit"]
    }
  },
  {
    name: "categoryAdd",
    path: "/categories/add",
    component: () =>
      import("@/pages/Entity/Category/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "categoryList", "categoryAdd"]
    }
  },
  /* News */
  {
    name: "newsList",
    path: "/news",
    component: () =>
      import("@/pages/Entity/News/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "newsList"]
    }
  },
  {
    name: "newsEdit",
    path: "/news/:id/edit",
    component: () =>
      import("@/pages/Entity/News/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "newsList", "newsEdit"]
    }
  },
  {
    name: "newsAdd",
    path: "/news/add",
    component: () =>
      import("@/pages/Entity/News/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "newsList", "newsAdd"]
    }
  },
  /* Tag */
  {
    name: "tagList",
    path: "/tags",
    component: () =>
      import("@/pages/Entity/Tag/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "tagList"]
    }
  },
  {
    name: "tagEdit",
    path: "/tags/:id/edit",
    component: () =>
      import("@/pages/Entity/Tag/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "tagList", "tagEdit"]
    }
  },
  {
    name: "tagAdd",
    path: "/tags/add",
    component: () =>
      import("@/pages/Entity/Tag/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "tagList", "tagAdd"]
    }
  },
  /* City */
  {
    name: "cityList",
    path: "/cities",
    component: () =>
      import("@/pages/Entity/City/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "cityList"]
    }
  },
  {
    name: "cityEdit",
    path: "/cities/:id/edit",
    component: () =>
      import("@/pages/Entity/City/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "cityList", "cityEdit"]
    }
  },
  {
    name: "cityAdd",
    path: "/cities/add",
    component: () =>
      import("@/pages/Entity/City/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "cityList", "cityAdd"]
    }
  },
  /* Country */
  {
    name: "countryList",
    path: "/countries",
    component: () =>
      import("@/pages/Entity/Country/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "countryList"]
    }
  },
  {
    name: "countryEdit",
    path: "/countries/:id/edit",
    component: () =>
      import("@/pages/Entity/Country/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "countryList", "countryEdit"]
    }
  },
  {
    name: "countryAdd",
    path: "/countries/add",
    component: () =>
      import("@/pages/Entity/Country/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "countryList", "countryAdd"]
    }
  },
  /* Region */
  {
    name: "regionList",
    path: "/regions",
    component: () =>
      import("@/pages/Entity/Region/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "regionList"]
    }
  },
  {
    name: "regionEdit",
    path: "/regions/:id/edit",
    component: () =>
      import("@/pages/Entity/Region/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "regionList", "regionEdit"]
    }
  },
  {
    name: "regionAdd",
    path: "/regions/add",
    component: () =>
      import("@/pages/Entity/Region/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "regionList", "regionAdd"]
    }
  },
  /* EncryptionKey */
  {
    name: "encryptionKeyList",
    path: "/encryption-keys",
    component: () =>
      import("@/pages/Entity/EncryptionKey/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "encryptionKeyList"]
    }
  },
  {
    name: "encryptionKeyEdit",
    path: "/encryption-keys/:id/edit",
    component: () =>
      import("@/pages/Entity/EncryptionKey/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "encryptionKeyList", "encryptionKeyEdit"]
    }
  },
  {
    name: "encryptionKeyAdd",
    path: "/encryption-keys/add",
    component: () =>
      import("@/pages/Entity/EncryptionKey/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "encryptionKeyList", "encryptionKeyAdd"]
    }
  },
  /* LoginCode */
  {
    name: "loginCodeList",
    path: "/login-codes",
    component: () =>
      import("@/pages/Entity/LoginCode/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "loginCodeList"]
    }
  },
  {
    name: "loginCodeEdit",
    path: "/login-codes/:id/edit",
    component: () =>
      import("@/pages/Entity/LoginCode/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "loginCodeList", "loginCodeEdit"]
    }
  },
  {
    name: "loginCodeAdd",
    path: "/login-codes/add",
    component: () =>
      import("@/pages/Entity/LoginCode/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "loginCodeList", "loginCodeAdd"]
    }
  },
  /* SiteUser */
  {
    name: "siteUserList",
    path: "/site-users",
    component: () =>
      import("@/pages/Entity/SiteUser/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "siteUserList"]
    }
  },
  {
    name: "siteUserEdit",
    path: "/site-users/:id/edit",
    component: () =>
      import("@/pages/Entity/SiteUser/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "siteUserList", "siteUserEdit"]
    }
  },
  {
    name: "siteUserAdd",
    path: "/site-users/add",
    component: () =>
      import("@/pages/Entity/SiteUser/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "siteUserList", "siteUserAdd"]
    }
  },
  /* VfsFile */
  {
    name: "vfsFileList",
    path: "/vfs-files",
    component: () =>
      import("@/pages/Entity/VfsFile/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "vfsFileList"]
    }
  },
  {
    name: "vfsFileEdit",
    path: "/vfs-files/:id/edit",
    component: () =>
      import("@/pages/Entity/VfsFile/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "vfsFileList", "vfsFileEdit"]
    }
  },
  {
    name: "vfsFileAdd",
    path: "/vfs-files/add",
    component: () =>
      import("@/pages/Entity/VfsFile/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "vfsFileList", "vfsFileAdd"]
    }
  },
  /* VfsFolder */
  {
    name: "vfsFolderList",
    path: "/vfs-folders",
    component: () =>
      import("@/pages/Entity/VfsFolder/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "vfsFolderList"]
    }
  },
  {
    name: "vfsFolderEdit",
    path: "/vfs-folders/:id/edit",
    component: () =>
      import("@/pages/Entity/VfsFolder/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "vfsFolderList", "vfsFolderEdit"]
    }
  },
  {
    name: "vfsFolderAdd",
    path: "/vfs-folders/add",
    component: () =>
      import("@/pages/Entity/VfsFolder/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "vfsFolderList", "vfsFolderAdd"]
    }
  },
];
