<template>
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
                    {{ t("news.list.title") }}
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
                <v-btn
                  small
                  dark
                  color="success"
                  :to="{ name: 'newsAdd' }"
                >
                  <v-icon>add</v-icon>
                  {{ t("common.list.addNewLabel") }}
                </v-btn>
              </v-flex>
            </v-layout>
          </v-flex>
        </v-layout>

        <!-- Complex Table -->
        <v-layout justify-center>
          <v-flex shrink>
            <v-card>
              <!-- Quick filter, chips -->
              <v-card-title class="pt-0">
                <v-layout
                  justify-space-between
                  align-end
                  wrap
                  class="flex-sm-nowrap"
                >
                  <v-flex
                    xs12
                    sm4
                    md3
                    mr-sm-2
                  >
                    <v-text-field
                      v-model="filters.title"
                      :placeholder="
                        t('news.list.filter.quickFilterPlaceholder')
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
                  </v-flex>
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

              <!-- Table -->
              <v-data-table
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
                    'min-width-table': $vuetify.breakpoint.mdAndUp,
                    'min-width-table-full': $vuetify.breakpoint.smAndDown,
                    smAndDown: $vuetify.breakpoint.smAndDown
                  }
                ]"
                :show-select="false"
                :loading="isLoading"
                fixed-header
                @update:options="setPagination"
              >
                <template #item.title="{ item }">
                  <router-link
                    :to="{ name: 'newsEdit', params: { id: item.id } }"
                    class="font-weight-medium"
                  >
                    {{ item.title }}
                  </router-link>
                </template>
                <template #item.preview="{ item }">
                  {{ item.preview }}
                </template>
                <template #item.content="{ item }">
                  {{ item.content }}
                </template>
                <template #item.category="{ item }">
                  {{ item.category | getField("title") }}
                </template>
                <template #item.country="{ item }">
                  {{ item.country | getField("title") }}
                </template>
                <template #item.region="{ item }">
                  {{ item.region | getField("title") }}
                </template>
                <template #item.status="{ item }">
                  <span class="text-no-wrap">
                    <vt-status-badge
                      v-model="item.status"
                      small
                    />
                  </span>
                </template>
                <template #item.id="{ item }">
                  <span class="text-no-wrap">
                    <v-hover v-slot="{ hover }">
                      <v-btn
                        text
                        dark
                        icon
                        :color="hover ? 'red' : 'grey'"
                        @click="deleteItem(item, 'title')"
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

<script lang="ts">
import { computed, defineComponent } from 'vue';
import { NewsSummary, NewsSearch } from '@/services/api/factory';
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
    } = useEntityList(News, NewsSearch);

    const headers = computed(() => [
      {
        text: t('news.list.headers.title'),
        value: 'title',
        align: 'left'
      },
      {
        text: t('news.list.headers.preview'),
        value: 'preview'
      },
      {
        text: t('news.list.headers.content'),
        value: 'content'
      },
      {
        text: t('news.list.headers.category'),
        value: 'category',
        sortable: false
      },
      {
        text: t('news.list.headers.country'),
        value: 'country',
        sortable: false
      },
      {
        text: t('news.list.headers.region'),
        value: 'region',
        sortable: false
      },
      {
        text: t('news.list.headers.status'),
        value: 'status',
        sortable: false
      },
      {
        text: t('news.list.headers.actions'),
        value: 'id',
        sortable: false
      }
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
