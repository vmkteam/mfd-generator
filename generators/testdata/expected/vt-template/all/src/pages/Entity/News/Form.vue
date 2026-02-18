<template>
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
              {{ model.title || "..." }}
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
                @click.stop="onDelete('title')"
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
                  <!--  generated part -->
                  <vt-form-field
                    v-model="model.title"
                    component="v-text-field"
                    :label="t('news.form.titleLabel')"
                    :error-messages="getErrorMessage(errors.title)"
                    :disabled="isLoading"
                    placeholder=""
                    required
                  /><vt-form-field
                    v-model="model.preview"
                    component="v-text-field"
                    :label="t('news.form.previewLabel')"
                    :error-messages="getErrorMessage(errors.preview)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                  /><vt-form-field
                    v-model="model.content"
                    component="vt-tinymce-editor"
                    :label="t('news.form.contentLabel')"
                    :error-messages="getErrorMessage(errors.content)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                    without-help
                  /><vt-form-field
                    v-model="model.categoryId"
                    entity="category"
                    search-by="title"
                    prefetch
                    component="vt-entity-autocomplete"
                    :label="t('news.form.categoryIdLabel')"
                    :error-messages="getErrorMessage(errors.categoryId)"
                    :disabled="isLoading"
                    placeholder=""
                    required
                  /><vt-form-field
                    v-model="model.countryId"
                    entity="country"
                    search-by="title"
                    prefetch
                    component="vt-entity-autocomplete"
                    :label="t('news.form.countryIdLabel')"
                    :error-messages="getErrorMessage(errors.countryId)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                  /><vt-form-field
                    v-model="model.regionId"
                    entity="region"
                    search-by="title"
                    prefetch
                    component="vt-entity-autocomplete"
                    :label="t('news.form.regionIdLabel')"
                    :error-messages="getErrorMessage(errors.regionId)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                  /><vt-form-field
                    v-model="model.cityId"
                    entity="city"
                    search-by="title"
                    prefetch
                    component="vt-entity-autocomplete"
                    :label="t('news.form.cityIdLabel')"
                    :error-messages="getErrorMessage(errors.cityId)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                  /><vt-form-field
                    v-model="model.tagIds"
                    entity="tag"
                    search-by="title"
                    prefetch
                    component="vt-entity-autocomplete"
                    :label="t('news.form.tagIdsLabel')"
                    :error-messages="getErrorMessage(errors.tagIds)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                    multiple
                    chips
                  /><vt-form-field
                    v-model="model.publishedAt"
                    component="vt-datetime-picker"
                    :label="t('news.form.publishedAtLabel')"
                    :error-messages="getErrorMessage(errors.publishedAt)"
                    :disabled="isLoading"
                    placeholder=""
                    clearable
                    iso
                  /><vt-form-field
                    v-model="model.statusId"
                    component="vt-status-select"
                    :label="t('news.form.statusIdLabel')"
                    :error-messages="getErrorMessage(errors.statusId)"
                    :disabled="isLoading"
                    placeholder=""
                    required
                    compact
                    :row="$vuetify.breakpoint.smAndUp"
                  /><!--  end generated part -->
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

<script lang="ts">
import { useEntityForm } from '@/composables/useEntityForm';
import { useI18n } from '@/composables/useI18n'
import { News as Model } from '@/services/api/factory';
import { defineComponent } from 'vue';

export default defineComponent({
  // eslint-disable-next-line vue/match-component-file-name
  name: 'NewsForm',

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
    } = useEntityForm<Model>(Model);

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
