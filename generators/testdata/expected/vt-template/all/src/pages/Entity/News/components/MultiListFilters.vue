<template>
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
    {
      id: 'title',
      type: 'input',
      title: t('news.list.filter.title'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        component: 'v-text-field'
      }
    },
    {
      id: 'preview',
      type: 'input',
      title: t('news.list.filter.preview'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        component: 'v-text-field'
      }
    },
    {
      id: 'content',
      type: 'input',
      title: t('news.list.filter.content'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        component: 'v-text-field'
      }
    },
    {
      id: 'categoryId',
      type: 'select',
      title: t('news.list.filter.categoryId'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        entity: 'category',
        searchBy: 'title',
        async: true,
        component: 'vt-entity-autocomplete'
      }
    },
    {
      id: 'countryId',
      type: 'select',
      title: t('news.list.filter.countryId'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        entity: 'country',
        searchBy: 'title',
        async: true,
        component: 'vt-entity-autocomplete'
      }
    },
    {
      id: 'regionId',
      type: 'select',
      title: t('news.list.filter.regionId'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        entity: 'region',
        searchBy: 'title',
        async: true,
        component: 'vt-entity-autocomplete'
      }
    },
    {
      id: 'cityId',
      type: 'select',
      title: t('news.list.filter.cityId'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        entity: 'city',
        searchBy: 'title',
        async: true,
        component: 'vt-entity-autocomplete'
      }
    },
    {
      id: 'createdAt',
      type: 'datetime',
      title: t('news.list.filter.createdAt'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        iso: true,
        component: 'vt-datetime-picker'
      }
    },
    {
      id: 'publishedAt',
      type: 'datetime',
      title: t('news.list.filter.publishedAt'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        iso: true,
        component: 'vt-datetime-picker'
      }
    },
    {
      id: 'statusId',
      type: 'select',
      title: t('news.list.filter.statusId'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        itemText: 'text',
        itemValue: 'value',
        component: 'vt-status-select'
      }
    },
    {
      type: 'divider'
    },
    {
      id: 'ids',
      type: 'select',
      title: t('news.list.filter.ids'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        itemText: 'text',
        itemValue: 'value',
        component: 'v-text-field'
      }
    }
  ].filter(Boolean) as IFilterItem[];

    return {
      t,
      filterItems
    };
  }
});
</script>
