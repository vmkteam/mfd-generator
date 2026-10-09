<template>
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
                    {{ t("tag.list.title") }}
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
                <vt-btn
                  small
                  dark
                  color="success"
                  :to="{ name: 'tagAdd' }"
                >
                  <vt-icon>mdi-plus</vt-icon>
                  {{ t("common.list.addNewLabel") }}
                </vt-btn>
              </vt-col>
            </vt-row>
          </vt-col>
        </vt-row>

        <!-- Complex Table -->
        <vt-row justify="center">
          <vt-col shrink>
            <vt-card>
              <!-- Quick filter, chips -->
              <v-card-title class="pt-0">
                <vt-row
                  justify="space-between"
                  align="end"
                  class="flex-wrap flex-sm-nowrap"
                >
                  <vt-col
                    cols="12"
                    sm="4"
                    md="3"
                    class="mr-sm-2"
                  >
                    <v-text-field
                      v-model="filters.title"
                      :placeholder="
                        t('tag.list.filter.quickFilterPlaceholder')
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
                  </vt-col>
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

              <!-- Table -->
              <vt-data-table
                v-model="selected"
                :headers="headers"
                :items="list"
                :options="vuetifyTableOptions"
                :server-items-length="pagination.totalItems"
                item-key="id"
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
              >
                <template #item.title="{ item }">
                  <router-link
                    :to="{ name: 'tagEdit', params: { id: item.id } }"
                    class="font-weight-medium"
                  >
                    {{ item.title }}
                  </router-link>
                </template>
                <template #item.status="{ item }">
                  <span class="text-no-wrap">
                    <vt-status-badge
                      :value="item.status"
                      small
                    />
                  </span>
                </template>
                <template #item.id="{ item }">
                  <span class="text-no-wrap">
                    <vt-hover v-slot="{ hover, props }">
                      <vt-btn
                        v-bind="props"
                        text
                        dark
                        icon
                        :color="hover ? 'red' : 'grey'"
                        @click="deleteItem(item, 'title')"
                      >
                        <vt-icon small>mdi-delete</vt-icon>
                      </vt-btn>
                    </vt-hover>
                  </span>
                </template>
              </vt-data-table>
            </vt-card>
          </vt-col>
        </vt-row>
      </vt-col>
    </vt-row>
  </vt-entity-view>
</template>

<script lang="ts">
import { useEntityList } from '@/composables/useEntityList';
import { useI18n } from '@/composables/useI18n';
import { Tag, TagSearch } from '@/services/api/factory';
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
    } = useEntityList(Tag, TagSearch);

    const headers = computed(() => [
      {
        text: t('tag.list.headers.title'),
        align: 'left',
        value: 'title'
      },
      {
        text: t('tag.list.headers.status'),
        align: 'left',
        value: 'status',
        sortable: false
      },
      {
        text: t('tag.list.headers.actions'),
        value: 'id',
        align: 'right',
        sortable: false
      }
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
