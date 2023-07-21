package vttmpl

const routesDefaultTemplate = `/* eslint-disable */
export default [{{range $model := .Entities}}
  /* {{.Name}} */
  {
    name: "{{.JSName}}List",
    path: "/{{.TerminalPath}}",
    component: () =>
      import("@/pages/Entity/{{.Name}}/List.vue"),
    meta: {
      breadcrumbs: ["dashboard", "{{.JSName}}List"]
    }
  },
  {{if not .ReadOnly}}{
    name: "{{.JSName}}Edit",
    path: "/{{.TerminalPath}}/:id/edit",
    component: () =>
      import("@/pages/Entity/{{.Name}}/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "{{.JSName}}List", "{{.JSName}}Edit"]
    }
  },
  {
    name: "{{.JSName}}Add",
    path: "/{{.TerminalPath}}/add",
    component: () =>
      import("@/pages/Entity/{{.Name}}/Form.vue"),
    meta: {
      breadcrumbs: ["dashboard", "{{.JSName}}List", "{{.JSName}}Add"]
    }
  },{{end}}{{end}}
];
`

// fuck backtick js
const listDefaultTemplate = `<template>
  <vt-entity-view>
    <v-layout
      align-start
      justify-center
    >
      <v-flex column>
        <v-layout justify-center>
          <v-flex
            xs12
            md8
          >
            <v-layout
              align-center
              mb-2
              wrap
            >
              <v-flex>
                <v-layout align-center>
                  <h2 class="ellipsed mr-1">
                    {{ $t("[[.JSName]].list.title") }}
                  </h2>
                  <span
                    v-if="store.pagination.totalItems"
                    class="text--secondary subtitle-2"
                  >
                    {{ store.pagination.totalItems }}
                  </span>
                </v-layout>
              </v-flex>
              <v-spacer />
              <v-flex shrink>
                [[if not .ReadOnly]]<v-btn
                  dark
                  color="success"
                  :to="{ name: '[[.JSName]]Add' }"
                >
                  <v-icon left>
                    add
                  </v-icon>
                  {{ $t("common.list.addNewLabel") }}
                </v-btn>[[end]]
              </v-flex>
            </v-layout>
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
                  align-end
                  wrap
                  class="flex-sm-nowrap"
                >[[if .HasQuickFilter]]
                  <v-flex
                    xs12
                    sm4
                    md3
                    mr-sm-2
                  >
                    <v-text-field
                      v-model="store.filters.[[.TitleField]]"
                      :placeholder="
                        $t('[[.JSName]].list.filter.quickFilterPlaceholder')
                      "
                      hide-details
                      @keyup.enter.native="submitFilters()"
                    />
                  </v-flex>
                  <v-flex
                    xs12
                    ml-sm-10
                    mr-sm-10
                  >
                    <multi-filters
                      :filters="store.filters"
                      :active-filters="store.activeFilters"
                      @submitFilters="submitFilters"
                    />
                  </v-flex>[[end]]
                  <v-flex
                    v-if="$vuetify.breakpoint.smAndUp"
                    mt-sm-4
                  >
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
                v-model="selected"
                :headers="headers"
                :items="store.list"
                :options="store.vuetifyTableOptions"
                :server-items-length="store.pagination.totalItems"
                item-key="[[range .PKs]][[.JSName]][[end]]"
                :footer-props="{
                  itemsPerPageOptions: [10, 25, 50, 100, 500]
                }"
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
                @update:options="setPagination"
              >[[range .ListColumns]][[if eq .JSName "statusId"]]
                <template #item.status="{ item }">
                  <span class="text-no-wrap">
                    <vt-status-badge
                      v-model="item.status"
                      small
                    />
                  </span>
                </template>[[else]]
                [[raw "<"]]template #item.[[.JSName]]="{ item }">[[if .IsBool]]
                  <vt-boolean-badge
                    :value="item.[[.JSName]]"
                    small
                  />[[else]][[if .EditLink]]
                  <router-link
                    :to="{ name: '[[$.JSName]]Edit', params: { [[range $.PKs]]id: item.[[.JSName]][[end]] } }"
                    class="font-weight-medium"
                  >[[end]]
                  [[if .EditLink]]  [[end]]{{ item.[[.JSName]][[if .HasPipe]] | [[.Pipe]][[end]] }}[[end]][[if .EditLink]]
                  </router-link>[[end]]
                </template>[[end]][[end]][[if not $.ReadOnly]]
                <template #item.[[range $.PKs]][[.JSName]]="{ item }"[[end]]>
                  <span class="text-no-wrap">
                    <v-hover v-slot="{ hover }">
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
                </template>[[end]]
              </v-data-table>
            </v-card>
          </v-flex>
        </v-layout>
      </v-flex>
    </v-layout>
  </vt-entity-view>
</template>

[[raw "<"]]script lang="ts">
import { Component } from 'vue-property-decorator';
import { Observer } from 'mobx-vue';
import EntityList from '@/common/Entity/EntityList';
import Store from '@/common/Entity/EntityCollectionStore';
import {
  [[.Name]]Summary as Model,
  [[.Name]]Search as SearchModel
} from '@/services/api/factory';
import MultiFilters from './components/MultiListFilters.vue';

@Observer
@Component({
  name: 'List',
  components: { MultiFilters }
})
export default class List extends EntityList {
  store: Store = new Store(Model, SearchModel);

  get headers () {
    return [
      [[range $i, $e := .ListColumns]][[if ne $i 0]]
      },
      [[end]]{[[if eq .JSName "statusId"]]
        text: this.$t('[[$.JSName]].list.headers.status'),
        value: 'status',
        sortable: false
      [[- else ]]
        text: this.$t('[[$.JSName]].list.headers.[[.JSName]]'),
        value: '[[.JSName]]'[[if eq $i 0]],
        align: 'left'[[end]][[if not .IsSortable]],
        sortable: false[[end]][[end]][[end]]
      [[- if $.ReadOnly ]]
      }
      [[- else ]]
      },
      {
        text: this.$t('[[$.JSName]].list.headers.actions'),
        value: 'id',
        sortable: false
      }[[end]]
    ];
  }
}
</script>

<style lang="scss"></style>
`

const filterDefaultTemplate = `<template>
  <vt-multi-filter
    :items="filterItems"
    :filters="filters"
    autofocus
    :label="$t('common.list.filter.title')"
    @submitFilters="$emit('submitFilters')"
  />
</template>

<script lang="ts">
import { Component } from 'vue-property-decorator';
import { Observer } from 'mobx-vue';
import EntityListFilters from '@/common/Entity/EntityListFilters';

@Observer
@Component
export default class MultiListFilters extends EntityListFilters {
  filterItems = [
[[- $filtersLen := len .FilterColumns ]]
[[- range $i, $e := .FilterColumns ]]
    [[- if (isLast $i $filtersLen) ]]
    {
      type: 'divider'
    },
    [[- end ]]
    {
      id: '[[ .JSName ]]',
      type: '[[ .SearchType ]]',
      title: this.$t('[[ $.JSName ]].list.filter.[[ .JSName ]]'),
      value: [[ if .IsCheckBox ]]true[[ else ]]null[[ end ]],
      values: null,
      settings: {
        placeholder: '',
        [[- if and .IsNumber ( eq .SearchType "input" ) ]]
        type: 'number',
        [[- end ]]
        [[- if and (or .IsCheckBox ( eq .SearchType "select")) (not .IsFK) ]]
        itemText: 'text',
        itemValue: 'value',
        [[- end ]]
        [[- if .IsFK ]]
        entity: '[[ .FKJSName | ToLower ]]',
        searchBy: '[[ .FKJSSearch ]]',
        async: true,
        [[- end ]]
        [[- if and .IsFK .IsArray ]]
        multiple: true,
        itemText: ' [[.FKJSSearch ]]',
        searchAdditional: {},
        [[- end ]]
        [[- if eq .Component "vt-datetime-picker" ]]
        iso: true,
        [[- end ]]
        component: '[[ .Component ]]'
      }
    }[[ if (notLast $i $filtersLen) ]],[[ end ]]
[[- end ]]
  ].filter(Boolean)
}
</script>
`

const formDefaultTemplate = `<template>
  <vt-entity-view>
    <v-layout
      align-start
      justify-center
    >
      <v-flex
        xs12
        md8
        mb-2
      >
        <v-layout
          mb-2
          wrap
        >
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
              <v-icon
                :left="!$vuetify.breakpoint.xsOnly"
              >
                arrow_back
              </v-icon>
              <template v-if="!$vuetify.breakpoint.xsOnly">
                {{ $t("common.form.cancelButtonLabel") }}
              </template>
            </v-btn>
            <v-hover
              v-if="$route.params.id"
              v-slot="{ hover }"
            >
              <v-btn
                :color="hover ? 'error' : ''"
                icon
                @click.stop="onDelete('[[.TitleField]]')"
              >
                <v-icon>delete</v-icon>
              </v-btn>
            </v-hover>
          </v-flex>
        </v-layout>
        <v-tabs
          v-model="tab"
          mobile-break-point="0"
        >
          <v-tab
            :class="{
              'error--text': tabsHasError.includes(0)
            }"
          >
            Основные
          </v-tab>
        </v-tabs>
        <v-card v-if="store.model">
          <v-form
            ref="form"
            @submit.prevent="onSaveAndBack"
          >
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
                    v-model="store.model.[[.JSName]]"[[if .IsFK]]
                    entity="[[ .FKJSName | ToLower ]]"
                    search-by="[[.FKJSSearch]]"
                    prefetch[[end]]
                    component="[[.Component]]"
                    :label="$t('[[$.JSName]].form.[[.JSName]]Label')"
                    :error-messages="$t(i18nFieldError(store.errors.[[.JSName]]))"
                    :disabled="store.isLoading"
                    placeholder=""[[if .Required]]
                    required[[else]]
                    clearable[[end]][[if eq .Component "vt-datetime-picker"]]
                    iso[[end]][[range .Params]]
                    [[.]][[end]]
                  />[[end]][[end]][[raw "<!--  end generated part -->"]]
                </v-tab-item>
              </v-tabs-items>
            </v-card-text>
            <v-card-actions>
              <v-layout wrap>
                <v-flex
                  v-if="$vuetify.breakpoint.smAndUp"
                  xs3
                />
                <v-flex>
                  <v-layout wrap>
                    <v-btn
                      type="submit"
                      color="success"
                      :disabled="!store.isChanged || store.isLoading"
                      :loading="store.isLoading"
                      :block="$vuetify.breakpoint.xsOnly"
                      :class="!$vuetify.breakpoint.xsOnly && 'mx-2'"
                    >
                      <v-icon left>
                        done
                      </v-icon>
                      {{ $t("common.form.saveAndCloseButtonLabel") }}
                    </v-btn>

                    <v-btn
                      v-if="$route.params.id"
                      :disabled="!store.isChanged || store.isLoading"
                      :loading="store.isLoading"
                      :block="$vuetify.breakpoint.xsOnly"
                      :class="[
                        $vuetify.breakpoint.xsOnly && 'ml-0 mt-2',
                        $vuetify.breakpoint.smAndUp && 'ml-2'
                      ]"
                      outlined
                      color="accent"
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
import { Component } from 'vue-property-decorator';
import { Observer } from 'mobx-vue';
import { [[.Name]] as Model } from '@/services/api/factory';
import Store from '@/common/Entity/EntityModelStore';
import EntityForm from '@/common/Entity/EntityForm';

@Observer
@Component
export default class Form extends EntityForm {
  store: Store<Model> = new Store<Model>(Model);
}
</script>

<style scoped></style>
`
