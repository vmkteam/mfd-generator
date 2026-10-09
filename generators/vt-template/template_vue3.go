package vttmpl

// vue3 template set: Vue 3 + Vuetify 3 components built with defineComponent/setup
// and the same composables as the composition set.

// routesVue3Template repeats the composition routes as is
const routesVue3Template = routesCompositionTemplate

// filterVue3Template repeats the composition multi filter as is
const filterVue3Template = filterCompositionTemplate

const listVue3Template = `<template>
  <vt-entity-view>
    <vt-row
      align="start"
      justify="center"
    >
      <vt-col
        cols="12"
        class="d-flex flex-column"
      >
        <vt-row justify="center">
          <vt-col
            cols="12"
            md="8"
          >
            <vt-row
              align="center"
              class="mb-2 flex-wrap"
            >
              <vt-col>
                <vt-row align="center">
                  <h2 class="ellipsed mr-1">
                    {{ t("[[.JSName]].list.title") }}
                  </h2>
                  <span
                    v-if="pagination.totalItems"
                    class="text-medium-emphasis text-subtitle-2"
                  >
                    {{ pagination.totalItems }}
                  </span>
                </vt-row>
              </vt-col>
              <v-spacer />
              <vt-col shrink>
                [[if not .ReadOnly]]<vt-btn
                  small
                  dark
                  color="success"
                  :to="{ name: '[[.JSName]]Add' }"
                >
                  <vt-icon>mdi-plus</vt-icon>
                  {{ t("common.list.addNewLabel") }}
                </vt-btn>[[end]]
              </vt-col>
            </vt-row>
          </vt-col>
        </vt-row>

        [[raw "<!-- Complex Table -->"]]
        <vt-row justify="center">
          <vt-col shrink>
            <vt-card>
              [[raw "<!-- Quick filter, chips -->"]]
              <v-card-title class="pt-0">
                <vt-row
                  justify="space-between"
                  align="end"
                  class="flex-wrap flex-sm-nowrap"
                >[[if .HasQuickFilter]]
                  <vt-col
                    cols="12"
                    sm="4"
                    md="3"
                    class="mr-sm-2"
                  >
                    <v-text-field
                      v-model="filters.[[.TitleField]]"
                      :placeholder="
                        t('[[.JSName]].list.filter.quickFilterPlaceholder')
                      "
                      hide-details
                      @keyup.enter="submitFilters()"
                    />
                  </vt-col>
                  <vt-col class="ml-sm-10 mr-sm-10">
                    <multi-filters
                      :filters="filters"
                      :active-filters="activeFilters"
                      @submitFilters="submitFilters"
                      @update:filters="updateFilters"
                    />
                  </vt-col>[[end]]
                  <vt-col
                    v-if="$vuetify.display.smAndUp"
                    shrink
                    class="mt-sm-4"
                  >
                    <vt-compact-pagination
                      :model-value="pagination.page"
                      :total-pages="pagination.totalPages"
                      @update:model-value="setCompactPagination"
                    />
                  </vt-col>
                </vt-row>
              </v-card-title>

              [[raw "<!-- Table -->"]]
              <vt-data-table
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
                    'min-width-table': $vuetify.display.mdAndUp,
                    'min-width-table-full': $vuetify.display.smAndDown,
                    smAndDown: $vuetify.display.smAndDown
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
                      :value="item.status"
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
                  [[if .EditLink]]  [[end]]{{ [[.Value]] }}[[end]][[if .EditLink]]
                  </router-link>[[end]]
                </template>[[end]][[end]][[if not $.ReadOnly]]
                <template #item.[[range $.PKs]][[.JSName]]="{ item }"[[end]]>
                  <span class="text-no-wrap">
                    <vt-hover v-slot="{ hover, props }">
                      <vt-btn
                        v-bind="props"
                        text
                        dark
                        icon
                        :color="hover ? 'red' : 'grey'"
                        @click="deleteItem(item, '[[.TitleField]]')"
                      >
                        <vt-icon small>mdi-delete</vt-icon>
                      </vt-btn>
                    </vt-hover>
                  </span>
                </template>[[end]]
              </vt-data-table>
            </vt-card>
          </vt-col>
        </vt-row>
      </vt-col>
    </vt-row>
  </vt-entity-view>
</template>

[[raw "<"]]script lang="ts">
import { useEntityList } from '@/composables/useEntityList';
import { useI18n } from '@/composables/useI18n';[[if .UsesTableDate]]
import { tableDate } from '@/helpers/date';[[end]]
import { [[.Name]], [[.Name]]Search } from '@/services/api/factory';
import { computed, defineComponent } from 'vue';

import MultiFilters from './components/MultiListFilters.vue';

export default defineComponent({
  name: 'List',

  components: { MultiFilters },

  setup () {
    const { t } = useI18n();

    const {
      list,
      filters,
      selected,
      isLoading,
      pagination,
      activeFilters,
      vuetifyTableOptions,

      deleteItem,
      submitFilters,
      setPagination,
      updateFilters,
      setCompactPagination
    } = useEntityList([[.Name]], [[.Name]]Search);

    const headers = computed(() => [
      [[range $i, $e := .ListColumns]][[if ne $i 0]]
      },
      [[end]]{[[if eq .JSName "statusId"]]
        text: t('[[$.JSName]].list.headers.status'),
        align: 'left',
        value: 'status',
        sortable: false
      [[- else ]]
        text: t('[[$.JSName]].list.headers.[[.JSName]]'),
        align: 'left',
        value: '[[.JSName]]'[[if not .IsSortable]],
        sortable: false[[end]][[end]][[end]]
      [[- if $.ReadOnly ]]
      }
      [[- else ]]
      },
      {
        text: t('[[$.JSName]].list.headers.actions'),
        value: 'id',
        align: 'right',
        sortable: false
      }[[end]]
    ]);

    return {
      t,
      list,
      headers,
      filters,
      selected,
      isLoading,
      pagination,
      activeFilters,
      vuetifyTableOptions,
[[if .UsesTableDate]]
      tableDate,[[end]]
      deleteItem,
      submitFilters,
      setPagination,
      updateFilters,
      setCompactPagination
    };
  }
});
</script>

<style lang="scss"></style>
`

const formVue3Template = `<template>
  <vt-entity-view>
    <vt-row
      align="start"
      justify="center"
    >
      <vt-col
        cols="12"
        md="8"
        class="mb-2"
      >
        <vt-row class="mb-2 flex-wrap">
          <vt-col>
            <h2 class="ellipsed">
              {{ model.[[.TitleField]] || "..." }}
            </h2>
          </vt-col>
          <v-spacer />
          <vt-col shrink>
            <vt-btn
              text
              color="primary"
              :disabled="isLoading"
              @click.stop="navigateBack"
            >
              <vt-icon
                :left="!$vuetify.display.xs"
              >
                mdi-arrow-left
              </vt-icon>
              <template v-if="!$vuetify.display.xs">
                {{ t("common.form.cancelButtonLabel") }}
              </template>
            </vt-btn>
            <vt-hover
              v-if="$route.params.id"
              v-slot="{ hover, props }"
            >
              <vt-btn
                v-bind="props"
                :color="hover ? 'error' : 'grey'"
                icon
                @click.stop="onDelete('[[.TitleField]]')"
              >
                <vt-icon>mdi-delete</vt-icon>
              </vt-btn>
            </vt-hover>
          </vt-col>
        </vt-row>
        <vt-card v-if="model">
          <v-tabs v-model="tab">
            <v-tab
              :class="{
                'text-error': tabsHasError.includes(0)
              }"
            >
              Основные
            </v-tab>
          </v-tabs>

          <v-form
            ref="form"
            @submit.prevent="onSaveAndBack"
          >
            <v-card-text>
              <v-tabs-window v-model="tab">
                <v-tabs-window-item eager>
                  [[raw "<!--  generated part -->"]][[range .FormColumns]]
                  [[if .IsCheckBox]]<vt-form-field v-model="model.[[.JSName]]">
                    <template #component-slot>
                      [[raw "<v-checkbox"]]
                        v-model="model.[[.JSName]]"
                        [[raw ":error-messages="]]"getErrorMessage(errors.[[.JSName]])"
                        :disabled="isLoading"
                        :label="t('[[$.JSName]].form.[[.JSName]]Label')"
                        color="primary"
                      />
                    </template>
                  </vt-form-field>[[else]][[raw "<vt-form-field"]]
                    v-model="model.[[.JSName]]"[[if .IsFK]]
                    entity="[[ .FKJSName | ToLower ]]"
                    search-by="[[.FKJSSearch]]"
                    prefetch[[end]][[if not .IsDefaultComponent]]
                    component="[[.Component]]"[[end]]
                    :label="t('[[$.JSName]].form.[[.JSName]]Label')"
                    :error-messages="getErrorMessage(errors.[[.JSName]])"
                    :disabled="isLoading"[[if eq .Component "vt-datetime-picker"]]
                    iso[[end]][[range .Params]]
                    [[.]][[end]][[if .Required]]
                    required[[else]]
                    clearable[[end]]
                  />[[end]][[end]]
                  [[raw "<!--  end generated part -->"]]
                </v-tabs-window-item>
              </v-tabs-window>
            </v-card-text>
            <v-card-actions>
              <vt-row class="flex-wrap">
                <vt-col
                  v-if="$vuetify.display.smAndUp"
                  cols="3"
                />
                <vt-col>
                  <vt-row class="flex-wrap">
                    <vt-btn
                      type="submit"
                      color="success"
                      :disabled="!isChanged || isLoading"
                      :loading="isLoading"
                      :block="$vuetify.display.xs"
                    >
                      <vt-icon>mdi-check</vt-icon>
                      {{ t("common.form.saveAndCloseButtonLabel") }}
                    </vt-btn>

                    <vt-btn
                      v-if="$route.params.id"
                      :disabled="!isChanged || isLoading"
                      :loading="isLoading"
                      :block="$vuetify.display.xs"
                      :class="[
                        $vuetify.display.xs && 'ml-0 mt-2',
                        $vuetify.display.smAndUp && 'ml-2'
                      ]"
                      @click.stop="onSave"
                    >
                      {{ t("common.form.saveButtonLabel") }}
                    </vt-btn>
                    <v-spacer />
                  </vt-row>
                </vt-col>
              </vt-row>
            </v-card-actions>
          </v-form>
        </vt-card>
      </vt-col>
    </vt-row>
  </vt-entity-view>
</template>

[[raw "<script"]] lang="ts">
import { useEntityForm } from '@/composables/useEntityForm';
import { useErrorMessage } from '@/composables/useErrorMessage';
import { useI18n } from '@/composables/useI18n';

import { [[.Name]] as Model } from '@/services/api/factory';
import { defineComponent } from 'vue';

export default defineComponent({
  name: '[[.Name]]Form',

  setup () {
    const { t } = useI18n();
    const { getErrorMessage } = useErrorMessage();

    const {
      tab,
      form,
      model,
      errors,
      isLoading,
      isChanged,
      tabsHasError,

      onSave,
      onDelete,
      navigateBack,
      onSaveAndBack
    } = useEntityForm({ Model });

    return {
      t,
      tab,
      form,
      model,
      errors,
      isLoading,
      isChanged,
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
