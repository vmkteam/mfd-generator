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
];
