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
      title: this.$t('news.list.filter.title'),
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
      title: this.$t('news.list.filter.preview'),
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
      title: this.$t('news.list.filter.content'),
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
      title: this.$t('news.list.filter.categoryId'),
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
      id: 'createdAt',
      type: 'datetime',
      title: this.$t('news.list.filter.createdAt'),
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
      title: this.$t('news.list.filter.publishedAt'),
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
      title: this.$t('news.list.filter.statusId'),
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
      title: this.$t('news.list.filter.ids'),
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
