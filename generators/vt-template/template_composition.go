package vttmpl

const routesCompositionTemplate = `/* eslint-disable */
export default [{{range $model := .Entities}}
    /* {{.Name}} */
    {
        name: "{{.JSName}}List",
        path: "/{{.TerminalPath}}",
        component: () =>
            import("@/pages/Entity/{{.Name}}/List.vue"),
        meta: {
            title: "{{.Name}}List",
            breadcrumbs: ["dashboard", "{{.JSName}}List"]
        }
    },
    {{if not .ReadOnly}}{
        name: "{{.JSName}}Edit",
        path: "/{{.TerminalPath}}/:id/edit",
        component: () =>
            import("@/pages/Entity/{{.Name}}/Form.vue"),
        meta: {
            title: "{{.Name}}Edit",
            breadcrumbs: ["dashboard", "{{.JSName}}List", "{{.JSName}}Edit"]
        }
    },
    {
        name: "{{.JSName}}Add",
        path: "/{{.TerminalPath}}/add",
        component: () =>
            import("@/pages/Entity/{{.Name}}/Form.vue"),
        meta: {
            title: "{{.Name}}Add",
            breadcrumbs: ["dashboard", "{{.JSName}}List", "{{.JSName}}Add"]
        }
    },{{end}}{{end}}
];
`

// fuck backtick js
const listCompositionTemplate = `<template>
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
                    {{ t("[[.JSName]].list.title") }}
                  </h2>
                  <span
                    v-if="pagination.totalItems"
                    class="text--secondary subtitle-2"
                  >
                    {{ pagination.totalItems }}
                  </span>
                </v-layout>
              </v-flex>
              <v-spacer />
              <v-flex shrink>
                [[if not .ReadOnly]]<v-btn
                  small
                  dark
                  color="success"
                  :to="{ name: '[[.JSName]]Add' }"
                >
                  <v-icon>add</v-icon>
                  {{ t("common.list.addNewLabel") }}
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
                      v-model="filters.[[.TitleField]]"
                      :placeholder="
                        t('[[.JSName]].list.filter.quickFilterPlaceholder')
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
                      :filters="filters"
                      :active-filters="activeFilters"
                      @submitFilters="submitFilters"
                      @update:filters="updateFilters"
                    />
                  </v-flex>[[end]]
                  <v-flex
                    v-if="$vuetify.breakpoint.smAndUp"
                    mt-sm-4
                  >
                    <vt-compact-pagination
                      :value="pagination.page"
                      :total-pages="pagination.totalPages"
                      @input="setCompactPagination"
                    />
                  </v-flex>
                </v-layout>
              </v-card-title>

              [[raw "<!-- Table -->"]]
              <v-data-table
                v-model="selected"
                :headers="headers"
                :items="list"
                :options="vuetifyTableOptions"
                :server-items-length="pagination.totalItems"
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
                :loading="isLoading"
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
import { computed, defineComponent } from 'vue';
import { [[.Name]]Summary, [[.Name]]Search } from '@/services/api/factory';
import { useEntityList } from '@/composables/useEntityList';
import { useI18n } from '@/composables/useI18n';

import MultiFilters from './components/MultiListFilters.vue';

export default defineComponent({
  // eslint-disable-next-line vue/multi-word-component-names
  name: 'List',
  components: { MultiFilters },

  setup () {
    const { t } = useI18n();

    const {
      selected,
      pagination,
      filters,
      activeFilters,
      list,
      isLoading,
      vuetifyTableOptions,
      deleteItem,
      submitFilters,
      setCompactPagination,
      setPagination,
      updateFilters
    } = useEntityList([[.Name]]Summary, [[.Name]]Search);

    const headers = computed(() => [
      [[range $i, $e := .ListColumns]][[if ne $i 0]]
      },
      [[end]]{[[if eq .JSName "statusId"]]
        text: t('[[$.JSName]].list.headers.status'),
        value: 'status',
        sortable: false
      [[- else ]]
        text: t('[[$.JSName]].list.headers.[[.JSName]]'),
        value: '[[.JSName]]'[[if eq $i 0]],
        align: 'left'[[end]][[if not .IsSortable]],
        sortable: false[[end]][[end]][[end]]
      [[- if $.ReadOnly ]]
      }
      [[- else ]]
      },
      {
        text: t('[[$.JSName]].list.headers.actions'),
        value: 'id',
        sortable: false
      }[[end]]
    ]);

    return {
      t,
      selected,
      pagination,
      filters,
      activeFilters,
      list,
      isLoading,
      vuetifyTableOptions,
      deleteItem,
      submitFilters,
      setCompactPagination,
      setPagination,
      updateFilters,
      headers
    };
  }
});
</script>

<style lang="scss"></style>
`

const filterCompositionTemplate = `<template>
  <vt-new-multi-filter
    :items="filterItems"
    :filters="filters"
    autofocus
    :label="t('common.list.filter.title')"
    @submitFilters="$emit('submitFilters')"
    @update:filters="$emit('update:filters', $event)"
  />
</template>

<script lang="ts">
import { IFilterItem } from '@/common/MultiFilter/types';
import { useI18n } from '@/composables/useI18n';
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'MultiListFilters',

  props: {
    filters: {
      type: Object,
      required: true
    },
    activeFilters: {
      type: Object,
      required: true
    }
  },

  emits: ['submitFilters', 'update:filters'],

  setup () {
    const { t } = useI18n();

    const filterItems = [
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
      title: t('[[ $.JSName ]].list.filter.[[ .JSName ]]'),
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
  ].filter(Boolean) as IFilterItem[];

    return {
      t,
      filterItems
    };
  }
});
</script>
`

const formCompositionTemplate = `<template>
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
              {{ model.[[.TitleField]] || "..." }}
            </h2>
          </v-flex>
          <v-spacer />
          <v-flex shrink>
            <v-btn
              text
              color="primary"
              :disabled="isLoading"
              @click.stop="navigateBack"
            >
              <v-icon
                :left="!$vuetify.breakpoint.xsOnly"
              >
                arrow_back
              </v-icon>
              <template v-if="!$vuetify.breakpoint.xsOnly">
                {{ t("common.form.cancelButtonLabel") }}
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
        <v-card v-if="model">
          <v-form
            ref="form"
            @submit.prevent="onSaveAndBack"
          >
            <v-card-text>
              <v-tabs-items v-model="tab">
                <v-tab-item eager>
                  [[raw "<!--  generated part -->"]]
                  [[range .FormColumns]][[if .IsCheckBox]]<vt-form-field v-model="model.[[.JSName]]">
					<template #component-slot>
	                  [[raw "<v-checkbox"]]
	                    v-model="model.[[.JSName]]"
						[[raw ":error-messages="]]"getErrorMessage(errors.[[.JSName]])"
                    	:disabled="isLoading"
						:label="t('[[$.JSName]].form.[[.JSName]]Label')"
						color="primary"
	                  />
					</template>
                  </vt-form-field>
                  [[else]][[raw "<vt-form-field"]]
                    v-model="model.[[.JSName]]"[[if .IsFK]]
                    entity="[[ .FKJSName | ToLower ]]"
                    search-by="[[.FKJSSearch]]"
                    prefetch[[end]]
                    component="[[.Component]]"
                    :label="t('[[$.JSName]].form.[[.JSName]]Label')"
                    :error-messages="getErrorMessage(errors.[[.JSName]])"
                    :disabled="isLoading"
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
                      :disabled="!isChanged || isLoading"
                      :loading="isLoading"
                      :block="$vuetify.breakpoint.xsOnly"
                      :class="!$vuetify.breakpoint.xsOnly && 'mx-2'"
                    >
                      <v-icon left>
                        done
                      </v-icon>
                      {{ t("common.form.saveAndCloseButtonLabel") }}
                    </v-btn>

                    <v-btn
                      v-if="$route.params.id"
                      :disabled="!isChanged || isLoading"
                      :loading="isLoading"
                      :block="$vuetify.breakpoint.xsOnly"
                      :class="[
                        $vuetify.breakpoint.xsOnly && 'ml-0 mt-2',
                        $vuetify.breakpoint.smAndUp && 'ml-2'
                      ]"
                      outlined
                      color="accent"
                      @click.stop="onSave"
                    >
                      {{ t("common.form.saveButtonLabel") }}
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
import { useEntityForm } from '@/composables/useEntityForm';
import { useI18n } from '@/composables/useI18n'
import { [[.Name]] as Model } from '@/services/api/factory';
import { defineComponent } from 'vue';

export default defineComponent({
  // eslint-disable-next-line vue/match-component-file-name
  name: '[[.Name]]Form',

  setup () {
    const { t } = useI18n();

    const {
      tab,
      form,
      model,
      errors,
      isLoading,
      isChanged,
      i18nFieldError,
      tabsHasError,

      onSave,
      onDelete,
      navigateBack,
      onSaveAndBack
    } = useEntityForm<Model>({ Model} );

	const getErrorMessage = (errorKey: string | null): string => {
      const errorMessage = i18nFieldError(errorKey);
      return errorMessage ? t(errorMessage) : '';
    };

    return {
      t,
      tab,
      form,
      model,
      errors,
      isLoading,
      isChanged,
      i18nFieldError,
      tabsHasError,
      getErrorMessage,

      onSave,
      onDelete,
      navigateBack,
      onSaveAndBack
    };
  }
});
</script>

<style scoped></style>
`
