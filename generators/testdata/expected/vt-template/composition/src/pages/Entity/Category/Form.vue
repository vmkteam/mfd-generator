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
                    :label="t('category.form.titleLabel')"
                    :error-messages="getErrorMessage(errors.title)"
                    :disabled="isLoading"
                    placeholder=""
                    required
                  /><vt-form-field
                    v-model="model.orderNumber"
                    component="v-text-field"
                    :label="t('category.form.orderNumberLabel')"
                    :error-messages="getErrorMessage(errors.orderNumber)"
                    :disabled="isLoading"
                    placeholder=""
                    required
                  /><vt-form-field
                    v-model="model.statusId"
                    component="vt-status-select"
                    :label="t('category.form.statusIdLabel')"
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
import { Category as Model } from '@/services/api/factory';
import { defineComponent } from 'vue';

export default defineComponent({
  // eslint-disable-next-line vue/match-component-file-name
  name: 'CategoryForm',

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
