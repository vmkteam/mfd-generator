<template>
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
    {
      id: 'title',
      type: 'input',
      title: this.$t('tag.list.filter.title'),
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
      title: this.$t('tag.list.filter.statusId'),
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
      id: 'divider'
    },
    {
      id: 'ids',
      type: 'select',
      title: this.$t('tag.list.filter.ids'),
      value: null,
      values: null,
      settings: {
        placeholder: '',
        itemText: 'text',
        itemValue: 'value',
        component: 'v-text-field'
      }
    }
  ].filter(Boolean)
}
</script>
