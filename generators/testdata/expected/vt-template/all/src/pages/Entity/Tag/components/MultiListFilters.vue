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
      title: t('tag.list.filter.title'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        component: 'v-text-field'
      }
    },
    {
      id: 'statusId',
      type: 'select',
      title: t('tag.list.filter.statusId'),
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
      title: t('tag.list.filter.ids'),
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
