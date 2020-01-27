package vttmpl

const routesTemplate = `/* eslint-disable */
export default [{{range $model := .Entities}}
  /* {{.Name}} */
  {
    name: "{{.JSName}}List",
    path: "/{{.TerminalPath}}",
    component: () =>
      import(/* webpackChunkName: "{{.Name}}List" */ "pages/Entity/{{.Name}}/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "{{.JSName}}List"]
    }
  },
  {
    name: "{{.JSName}}Edit",
    path: "/{{.TerminalPath}}/:id/edit",
    component: () =>
      import(/* webpackChunkName: "{{.Name}}Form" */ "pages/Entity/{{.Name}}/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "{{.JSName}}List", "{{.JSName}}Edit"]
    }
  },
  {
    name: "{{.JSName}}Add",
    path: "/{{.TerminalPath}}/add",
    component: () =>
      import(/* webpackChunkName: "{{.Name}}Form" */ "pages/Entity/{{.Name}}/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "{{.JSName}}List", "{{.JSName}}Add"]
    }
  },{{end}}
];
`

// fuck backtick js
const listTemplate = `<template>
  <vt-entity-view>
    <v-layout align-start justify-center>
      <v-flex column>
        <v-layout justify-center>
          <v-flex xs12 md8>
            <v-layout align-center mb-2 wrap>
              <v-flex>
                <v-layout align-center>
                  <h2 class="ellipsed mr-1">
                    {{ $t("[[.JSName]].list.title") }}
                  </h2>
                  <span
                    class="text--secondary subtitle-2"
                    v-if="store.pagination.totalItems"
                  >
                    {{ store.pagination.totalItems }}
                  </span>
                </v-layout>
              </v-flex>
              <v-spacer />
              <v-flex shrink>
                <v-btn
                  @click.stop="filtersIsOpen = !filtersIsOpen"
                  :color="
` + "                    `${store.activeFiltersCount ? 'teal' : 'grey'} lighten-1`" + `
                  "
                  class="mr-2"
                  text
                  small
                >
                  <v-icon>mdi-filter</v-icon>
                  {{
                    store.activeFiltersCount
` + "                      ? `(${store.activeFiltersCount})`" + `
                      : ""
                  }}
                </v-btn>
                <v-btn small dark color="success" :to="{ name: '[[.JSName]]Add' }">
                  <v-icon>add</v-icon>
                  {{ $t("common.list.addNewLabel") }}
                </v-btn>
              </v-flex>
            </v-layout>
            <filters
              v-if="filtersIsOpen"
              :filters="store.filters"
              :active-filters="store.activeFilters"
              @submitFilters="submitFilters"
            />
          </v-flex>
        </v-layout>

        [[raw "<!-- Complex Table -->"]]
        <v-layout justify-center>
          <v-flex shrink>
            <v-card>
              [[raw "<!-- Quick filter, chips -->"]]
              <v-card-title class="pt-0">
                <v-layout
                  justify-space-between
                  align-start
                  wrap
                  class="flex-sm-nowrap"
                >[[if .HasQuickFilter]]
                  <v-flex
                    v-if="!store.activeFilters.[[.TitleField]]"
                    xs12
                    sm4
                    md3
                    mr-sm-2
                  >
                    <v-text-field
                      :placeholder="
                        $t('[[.JSName]].list.filter.quickFilterPlaceholder')
                      "
                      hide-details
                      v-model="store.filters.[[.TitleField]]"
                      @keyup.enter.native="submitFilters()"
                    />
                  </v-flex>
                  <v-flex
                    xs12
                    mt-sm-4
                    :class="` + "` ${store.activeFilters.[[.TitleField]] ? 'sm12' : 'sm9'}`" + `"
                  >
                    <vt-filters-chips
                      entity="[[.JSName]]"
                      :filters="store.activeFilters"
                      @reset:filters="resetFiltersKey"
                    />
                  </v-flex>
                  [[end]]
                  <v-flex v-if="$vuetify.breakpoint.smAndUp" mt-sm-4>
                    <vt-compact-pagination
                      :value="store.pagination.page"
                      :total-pages="store.pagination.totalPages"
                      @input="setCompactPagination"
                    />
                  </v-flex>
                </v-layout>
              </v-card-title>

              [[raw "<!-- Table -->"]]
              <v-data-table
                :headers="headers"
                :items="store.list"
                :options="store.vuetifyTableOptions"
                :server-items-length="store.pagination.totalItems"
                v-model="selected"
                item-key="[[range .PKs]][[.JSName]][[end]]"
                :footer-props="{
                  itemsPerPageOptions: [10, 25, 50, 100, 500]
                }"
                @update:options="setPagination"
                :class="[
                  'data-table-wrapper-sticky-fix',
                  {
                    'min-width-table': $vuetify.breakpoint.mdAndUp,
                    'min-width-table-full': $vuetify.breakpoint.smAndDown,
                    smAndDown: $vuetify.breakpoint.smAndDown
                  }
                ]"
                :show-select="false"
                :loading="store.isLoading"
                fixed-header
              >[[range .ListColumns]][[if eq .JSName "statusId"]]
                <template #item.status="{ item }">
                  <span class="text-no-wrap">
                    <vt-status-badge v-model="item.status" small />
                  </span>
                </template>[[else]]
                [[raw "<"]]template #item.[[.JSName]]="{ item }">[[if .IsBool]]
				  <vt-boolean-badge :value="item.[[.JSName]]" small />[[else]][[if .EditLink]]
                  <router-link
                    :to="{ name: '[[$.JSName]]Edit', params: { [[range $.PKs]]id: item.[[.JSName]][[end]] } }"
                    class="font-weight-medium"
                  >[[end]]
                  {{ item.[[.JSName]][[if .HasPipe]] | [[.Pipe]][[end]] }}[[end]][[if .EditLink]]
                  </router-link>[[end]]
                </template>[[end]][[end]]
                <template #item.[[range $.PKs]][[.JSName]]="{ item }"[[end]]>
                  <span class="text-no-wrap">
                    <v-btn
                      text
                      dark
                      :to="{ name: '[[.JSName]]Edit', params: { [[range .PKs]]id: item.[[.JSName]][[end]] } }"
                      icon
                      color="primary"
                    >
                      <v-icon small>edit</v-icon>
                    </v-btn>
                    <v-hover #default="{ hover }">
                      <v-btn
                        text
                        dark
                        icon
                        :color="hover ? 'red' : 'grey'"
                        @click="deleteItem(item, '[[.TitleField]]')"
                      >
                        <v-icon small>delete</v-icon>
                      </v-btn>
                    </v-hover>
                  </span>
                </template>
              </v-data-table>
            </v-card>
          </v-flex>
        </v-layout>
      </v-flex>
    </v-layout>
  </vt-entity-view>
</template>

[[raw "<"]]script lang="ts">
import { Component } from "vue-property-decorator";
import { Observer } from "mobx-vue";
import EntityList from "common/Entity/EntityList";
import Store from "common/Entity/EntityCollectionStore";
import {
  [[.Name]]Summary as Model,
  [[.Name]]Search as SearchModel
} from "services/api/factory";
import Filters from "./components/ListFilters.vue";

@Observer
@Component({
  name: "List",
  components: { Filters }
})
export default class List extends EntityList {
  store: Store = new Store(Model, SearchModel);

  get headers() {
    return [{[[range $i, $e := .ListColumns]][[if eq .JSName "statusId"]]
        text: this.$t("[[$.JSName]].list.headers.status"),
        value: "status"
      },
      {[[else]]
        text: this.$t("[[$.JSName]].list.headers.[[.JSName]]"),
        value: "[[.JSName]]"[[if eq $i 0]],
        align: "left"[[end]]
      },
      {[[end]][[end]]
        text: this.$t("[[$.JSName]].list.headers.actions"),
        value: "id",
        sortable: false
      }
    ];
  }
}
</script>

<style lang="scss"></style>
`

const filterTemplate = `<template>
  <v-layout mb-2>
    <v-flex>
      <v-card>
        <v-form @submit.prevent="$emit('submitFilters')">
          <v-card-text class="pb-0">
            [[raw "<!-- generated part -->"]]
            [[range $i, $e := .FilterColumns]]<vt-form-field[[if not .IsShortFilter]]
              v-if="isFullFilter || activeFilters.[[.JSName]]"[[end]]
              component="[[.Component]]" 
              :label="$t('[[$.JSName]].list.filter.[[.JSName]]')"[[if .IsFK]]
              entity="[[ .FKJSName | ToLower ]]"
              searchBy="[[.FKJSSearch]]"
			  async[[end]]
              placeholder=""
              v-model="filters.[[.JSName]]"
              class="mb-2"
              hide-details
              clearable[[if .ShowShortFilterLabel]]
            >
              <v-layout v-if="!isFullFilter">
                <v-flex xs12>
                  <v-subheader class="mb-2 mt-2 pl-0 pl-sm-4">
                    <a href="#" @click.stop.prevent="isFullFilter = true">
                      {{ $t("common.list.filter.allFiltersLabel") }}
                    </a>
                  </v-subheader>
                </v-flex>
              </v-layout>
            </vt-form-field>[[else]]
            />[[end]]
            [[end]][[raw "<!-- generated part end -->"]]
          </v-card-text>
          <v-card-actions class="pa-4 pt-0">
            <v-flex offset-sm-3>
              <v-btn color="primary" type="submit">
                {{ $t("common.list.filter.submitButtonLabel") }}
              </v-btn>
            </v-flex>
          </v-card-actions>
        </v-form>
      </v-card>
    </v-flex>
  </v-layout>
</template>

<script lang="ts">
import { Component } from "vue-property-decorator";
import { Observer } from "mobx-vue";
import EntityListFilters from "common/Entity/EntityListFilters";

@Observer
@Component
export default class Filters extends EntityListFilters {}
</script>

<style scoped></style>
`

const formTemplate = `<template>
  <vt-entity-view>
    <v-layout align-start justify-center>
      <v-flex xs12 md8 mb-2>
        <v-layout mb-2 wrap>
          <v-flex>
            <h2 class="ellipsed">
              {{ store.model.[[.TitleField]] || "..." }}
            </h2>
          </v-flex>
          <v-spacer />
          <v-flex shrink>
            <v-btn
              text
              color="primary"
              :disabled="store.isLoading"
              @click.stop="navigateBack"
            >
              {{ $t("common.form.cancelButtonLabel") }}
            </v-btn>
            <v-hover v-if="$route.params.id" #default="{ hover }">
              <v-btn
                :color="hover ? 'error' : ''"
                @click.stop="onDelete('[[.TitleField]]')"
                icon
              >
                <v-icon>delete</v-icon>
              </v-btn>
            </v-hover>
          </v-flex>
        </v-layout>
        <v-card v-if="store.model">
          <v-tabs v-model="tab" mobile-break-point="0">
            <v-tab
              :class="{
                'error--text': tabsHasError.includes(0)
              }"
            >
              Основные
            </v-tab>
          </v-tabs>

          <v-form @submit.prevent="onSaveAndBack" ref="form">
            <v-card-text>
              <v-tabs-items v-model="tab">
                <v-tab-item eager>
                  [[raw "<!--  generated part -->"]]
				  [[range .FormColumns]][[if .IsCheckBox]]<vt-form-field v-model="store.model.[[.JSName]]">
					<template #component-slot>
					  [[raw "<v-checkbox"]]
					    v-model="store.model.[[.JSName]]"
						[[raw ":error-messages="]]"$t(i18nFieldError(store.errors.[[.JSName]]))"
                    	:disabled="store.isLoading"
						:label="$t('[[$.JSName]].form.[[.JSName]]Label')"
						color="primary"
					  />
					</template>
                  </vt-form-field>
				  [[else]][[raw "<vt-form-field"]]
                    component="[[.Component]]"
                    :label="$t('[[$.JSName]].form.[[if eq .JSName "statusId"]]status[[else]][[.JSName]][[end]]Label')"
                    v-model="store.model.[[.JSName]]"[[if .IsFK]]
					entity="[[ .FKJSName | ToLower ]]"
					searchBy="[[.FKJSSearch]]"
					prefetch[[end]]
                    :error-messages="$t(i18nFieldError(store.errors.[[.JSName]]))"
                    :disabled="store.isLoading"
                    placeholder=""[[if .Required]]
					required[[else]]
					clearable[[end]][[range .Params]]
                    [[.]][[end]]
                  />
				  [[end]][[end]]
                  <!-- end of generated part -->
                </v-tab-item>
              </v-tabs-items>
            </v-card-text>
            <v-card-actions>
              <v-layout wrap>
                <v-flex v-if="$vuetify.breakpoint.smAndUp" xs3> </v-flex>
                <v-flex>
                  <v-layout wrap>
                    <v-btn
                      type="submit"
                      color="success"
                      :disabled="!store.isChanged || store.isLoading"
                      :loading="store.isLoading"
                      :block="$vuetify.breakpoint.xsOnly"
                    >
                      <v-icon>done</v-icon>
                      {{ $t("common.form.saveAndCloseButtonLabel") }}
                    </v-btn>

                    <v-btn
                      :disabled="!store.isChanged || store.isLoading"
                      :loading="store.isLoading"
                      :block="$vuetify.breakpoint.xsOnly"
                      :class="$vuetify.breakpoint.xsOnly && ['ml-0', 'mt-2']"
                      @click.stop="onSave"
                    >
                      {{ $t("common.form.saveButtonLabel") }}
                    </v-btn>
                    <v-spacer />
                  </v-layout>
                </v-flex>
              </v-layout>
            </v-card-actions>
          </v-form>
        </v-card>
      </v-flex>
    </v-layout>
  </vt-entity-view>
</template>

[[raw "<script"]] lang="ts">
import { Component } from "vue-property-decorator";
import { Observer } from "mobx-vue";
import { [[.Name]] as Model } from "services/api/factory";
import Store from "common/Entity/EntityModelStore";
import EntityForm from "common/Entity/EntityForm";

@Observer
@Component
export default class Form extends EntityForm {
  store: Store<Model> = new Store<Model>(Model);
}
</script>

<style scoped></style>
`
