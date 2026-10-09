<template>
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
              {{ model.title || "..." }}
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
                @click.stop="onDelete('title')"
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
                  <!--  generated part -->
                  <vt-form-field
                    v-model="model.title"
                    :label="t('category.form.titleLabel')"
                    :error-messages="getErrorMessage(errors.title)"
                    :disabled="isLoading"
                    required
                  />
                  <vt-form-field
                    v-model="model.orderNumber"
                    :label="t('category.form.orderNumberLabel')"
                    :error-messages="getErrorMessage(errors.orderNumber)"
                    :disabled="isLoading"
                    required
                  />
                  <vt-form-field
                    v-model="model.statusId"
                    component="vt-status-select"
                    :label="t('category.form.statusIdLabel')"
                    :error-messages="getErrorMessage(errors.statusId)"
                    :disabled="isLoading"
                    compact
                    :row="$vuetify.display.smAndUp"
                    required
                  />
                  <!--  end generated part -->
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

<script lang="ts">
import { useEntityForm } from '@/composables/useEntityForm';
import { useErrorMessage } from '@/composables/useErrorMessage';
import { useI18n } from '@/composables/useI18n';

import { Category as Model } from '@/services/api/factory';
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'CategoryForm',

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
