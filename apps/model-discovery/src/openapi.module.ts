import { Controller, Get, Param, Post, Put, Query } from '@nestjs/common';
import { Module } from '@nestjs/common';
import {
  ApiBearerAuth,
  ApiBody,
  ApiConsumes,
  ApiOkResponse,
  ApiOperation,
  ApiProperty,
  ApiPropertyOptional,
  ApiQuery,
  ApiTags,
} from '@nestjs/swagger';
import { ModelsQueryDto } from './catalog/dto/models-query.dto';
import { UpdateDisplaySettingsDto } from './display-settings/dto/update-display-settings.dto';
import { UpsertProviderConnectionProfileDto } from './providers/dto/upsert-provider-connection-profile.dto';
import { UpsertProviderDto } from './providers/dto/upsert-provider.dto';
import { UpsertModelDto } from './catalog/dto/upsert-model.dto';
import { UpsertModelSeriesDto } from './catalog/dto/upsert-model-series.dto';

class BooleanToggleDto {
  @ApiProperty()
  enabled!: boolean;
}

class IndexOneDto {
  @ApiProperty()
  modelId!: string;
}

class UploadFileDto {
  @ApiProperty({ type: 'string', format: 'binary' })
  file!: string;
}

class SimpleAckResponse {
  @ApiProperty()
  ok!: boolean;
}

class ExchangeRateSnapshot {
  @ApiProperty()
  usdToIrr!: number;

  @ApiProperty()
  fetchedAt!: string;

  @ApiProperty()
  source!: string;
}

class DisplaySettingsResponse {
  @ApiProperty({ enum: ['USD', 'IRR', 'TOMAN'] })
  currencyUnit!: 'USD' | 'IRR' | 'TOMAN';

  @ApiProperty({ enum: ['GREGORIAN', 'SHAMSI'] })
  dateCalendar!: 'GREGORIAN' | 'SHAMSI';

  @ApiProperty({ enum: ['PER_MILLION_TOKENS', 'PER_TOKEN'] })
  priceScale!: 'PER_MILLION_TOKENS' | 'PER_TOKEN';

  @ApiProperty({ type: () => ExchangeRateSnapshot })
  exchangeRate!: ExchangeRateSnapshot;
}

class FacetOption {
  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;
}

class FacetsResponse {
  @ApiProperty({ type: () => [FacetOption] })
  providers!: FacetOption[];

  @ApiProperty({ type: () => [FacetOption] })
  categories!: FacetOption[];

  @ApiProperty({ type: () => [FacetOption] })
  capabilities!: FacetOption[];

  @ApiProperty({ type: () => [FacetOption] })
  supportedParameters!: FacetOption[];

  @ApiProperty({ type: () => [FacetOption] })
  inputModalities!: FacetOption[];

  @ApiProperty({ type: () => [FacetOption] })
  outputModalities!: FacetOption[];

  @ApiProperty({ type: () => [FacetOption] })
  regions!: FacetOption[];
}

class ProviderConnectionProfileResponse {
  @ApiProperty()
  providerType!: string;

  @ApiPropertyOptional()
  authMode?: string | null;

  @ApiProperty({ additionalProperties: true })
  configJson!: Record<string, unknown>;
}

class ProviderCatalogEntry {
  @ApiProperty()
  id!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiPropertyOptional()
  description?: string | null;

  @ApiPropertyOptional()
  websiteUrl?: string | null;

  @ApiPropertyOptional()
  docsUrl?: string | null;

  @ApiPropertyOptional()
  logoUrl?: string | null;

  @ApiPropertyOptional()
  status?: string | null;

  @ApiProperty()
  enabled!: boolean;

  @ApiPropertyOptional()
  supportsByok?: boolean | null;

  @ApiPropertyOptional({ type: () => ProviderConnectionProfileResponse })
  connectionProfile?: ProviderConnectionProfileResponse | null;
}

class ProviderSummary {
  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiPropertyOptional()
  logoUrl?: string | null;
}

class ModelSeriesSummary {
  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiPropertyOptional()
  logoUrl?: string | null;
}

class CategorySummary {
  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;
}

class CapabilityLink {
  @ApiProperty({ type: () => FacetOption })
  capability!: FacetOption;
}

class SupportedParameterLink {
  @ApiProperty({ type: () => FacetOption })
  supportedParameter!: FacetOption;
}

class ModalityLink {
  @ApiProperty({ type: () => FacetOption })
  modality!: FacetOption;
}

class RegionLink {
  @ApiProperty({ type: () => FacetOption })
  region!: FacetOption;
}

class ModelPricingSnapshot {
  @ApiPropertyOptional()
  currency?: string | null;

  @ApiPropertyOptional()
  inputPricePerMillionTokens?: number | null;

  @ApiPropertyOptional()
  outputPricePerMillionTokens?: number | null;

  @ApiPropertyOptional()
  cachedInputPricePerMillionTokens?: number | null;

  @ApiPropertyOptional()
  imageInputPrice?: number | null;

  @ApiPropertyOptional()
  imageOutputPrice?: number | null;

  @ApiPropertyOptional()
  audioInputPrice?: number | null;

  @ApiPropertyOptional()
  audioOutputPrice?: number | null;

  @ApiPropertyOptional()
  requestPrice?: number | null;

  @ApiPropertyOptional()
  sourceUrl?: string | null;

  @ApiPropertyOptional()
  effectiveFrom?: string | null;

  @ApiPropertyOptional({ type: () => FacetOption })
  region?: FacetOption | null;
}

class ModelBenchmarkResult {
  @ApiPropertyOptional()
  benchmarkId?: string | null;

  @ApiPropertyOptional()
  score?: number | null;

  @ApiPropertyOptional()
  measuredAt?: string | null;
}

class ModelSourceReference {
  @ApiPropertyOptional()
  title?: string | null;

  @ApiProperty()
  url!: string;

  @ApiPropertyOptional()
  sourceType?: string | null;
}

class ModelChangelogEntry {
  @ApiPropertyOptional()
  changeType?: string | null;

  @ApiPropertyOptional()
  title?: string | null;

  @ApiPropertyOptional()
  description?: string | null;

  @ApiPropertyOptional()
  changedAt?: string | null;

  @ApiPropertyOptional()
  createdBy?: string | null;
}

class CatalogModelListItem {
  @ApiProperty()
  id!: string;

  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiProperty({ type: () => ProviderSummary })
  provider!: ProviderSummary;

  @ApiPropertyOptional({ type: () => ModelSeriesSummary })
  modelSeries?: ModelSeriesSummary | null;

  @ApiPropertyOptional()
  description?: string | null;

  @ApiPropertyOptional()
  status?: string | null;

  @ApiProperty({ type: () => [String] })
  mode!: string[];

  @ApiPropertyOptional()
  contextWindow?: number | null;

  @ApiPropertyOptional()
  maxOutputTokens?: number | null;

  @ApiProperty({ type: () => [String] })
  inputModalities!: string[];

  @ApiProperty({ type: () => [String] })
  outputModalities!: string[];

  @ApiProperty({ type: () => [String] })
  supportedParameters!: string[];

  @ApiProperty({ type: () => [String] })
  capabilities!: string[];

  @ApiPropertyOptional({ type: () => ModelPricingSnapshot })
  pricing?: ModelPricingSnapshot | null;

  @ApiProperty({ type: () => [String] })
  regions!: string[];

  @ApiProperty()
  openSource!: boolean;

  @ApiProperty()
  free!: boolean;

  @ApiProperty()
  createdAt!: string;

  @ApiProperty()
  updatedAt!: string;
}

class CatalogListPagination {
  @ApiProperty()
  page!: number;

  @ApiProperty()
  limit!: number;

  @ApiProperty()
  total!: number;

  @ApiProperty()
  totalPages!: number;
}

class CatalogListResponse {
  @ApiProperty({ type: () => [CatalogModelListItem] })
  data!: CatalogModelListItem[];

  @ApiProperty({ type: () => CatalogListPagination })
  pagination!: CatalogListPagination;
}

class CatalogModelDetailResponse {
  @ApiProperty()
  id!: string;

  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  displayName!: string;

  @ApiProperty({ type: () => ProviderCatalogEntry })
  provider!: ProviderCatalogEntry;

  @ApiPropertyOptional({ type: () => ModelSeriesSummary })
  modelSeries?: ModelSeriesSummary | null;

  @ApiPropertyOptional({ type: () => CategorySummary })
  category?: CategorySummary | null;

  @ApiPropertyOptional()
  shortDescription?: string | null;

  @ApiPropertyOptional()
  longDescription?: string | null;

  @ApiPropertyOptional()
  status?: string | null;

  @ApiPropertyOptional()
  lifecycleStatus?: string | null;

  @ApiProperty({ type: () => [String] })
  mode!: string[];

  @ApiPropertyOptional()
  contextWindow?: number | null;

  @ApiPropertyOptional()
  maxOutputTokens?: number | null;

  @ApiProperty({ type: () => [ModalityLink] })
  inputModalities!: ModalityLink[];

  @ApiProperty({ type: () => [ModalityLink] })
  outputModalities!: ModalityLink[];

  @ApiProperty({ type: () => [SupportedParameterLink] })
  supportedParams!: SupportedParameterLink[];

  @ApiProperty({ type: () => [CapabilityLink] })
  capabilities!: CapabilityLink[];

  @ApiProperty({ type: () => [RegionLink] })
  regions!: RegionLink[];

  @ApiProperty({ type: () => [ModelPricingSnapshot] })
  pricing!: ModelPricingSnapshot[];

  @ApiProperty({ type: () => [ModelSourceReference] })
  sources!: ModelSourceReference[];

  @ApiProperty({ type: () => [ModelChangelogEntry] })
  changelog!: ModelChangelogEntry[];

  @ApiProperty({ type: () => [ModelBenchmarkResult] })
  benchmarks!: ModelBenchmarkResult[];

  @ApiProperty()
  openSource!: boolean;

  @ApiProperty()
  free!: boolean;
}

class ProviderDetailModel {
  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  displayName!: string;

  @ApiProperty({ type: () => [String] })
  mode!: string[];

  @ApiPropertyOptional()
  shortDescription?: string | null;

  @ApiPropertyOptional()
  status?: string | null;

  @ApiProperty({ type: () => [ModelPricingSnapshot] })
  pricing!: ModelPricingSnapshot[];

  @ApiProperty({ type: () => [CapabilityLink] })
  capabilities!: CapabilityLink[];
}

class ProviderDetailResponse extends ProviderCatalogEntry {
  @ApiProperty({ type: () => [ProviderDetailModel] })
  models!: ProviderDetailModel[];
}

class ProviderModelsListItem {
  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiProperty({ type: () => [String] })
  mode!: string[];

  @ApiPropertyOptional()
  contextWindow?: number | null;

  @ApiPropertyOptional()
  maxOutputTokens?: number | null;

  @ApiProperty({ type: () => [String] })
  inputModalities!: string[];

  @ApiProperty({ type: () => [String] })
  outputModalities!: string[];

  @ApiProperty({ type: () => [String] })
  supportedParameters!: string[];

  @ApiProperty({ type: () => [String] })
  capabilities!: string[];

  @ApiPropertyOptional({ type: () => ModelPricingSnapshot })
  pricing?: ModelPricingSnapshot | null;

  @ApiProperty({ type: () => [String] })
  regions!: string[];
}

class ProviderModelsListResponse {
  @ApiProperty({ type: () => [ProviderModelsListItem] })
  data!: ProviderModelsListItem[];
}

class AuditLogEntry {
  @ApiProperty()
  id!: string;

  @ApiProperty()
  entityType!: string;

  @ApiProperty()
  entityId!: string;

  @ApiProperty()
  action!: string;

  @ApiPropertyOptional()
  actorId?: string | null;

  @ApiPropertyOptional({ additionalProperties: true })
  beforeJson?: Record<string, unknown> | null;

  @ApiPropertyOptional({ additionalProperties: true })
  afterJson?: Record<string, unknown> | null;

  @ApiProperty()
  createdAt!: string;
}

class ModelSeriesRecord {
  @ApiProperty()
  id!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiPropertyOptional()
  logoUrl?: string | null;
}

class AdminModelRecord {
  @ApiProperty()
  id!: string;

  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  displayName!: string;

  @ApiPropertyOptional()
  status?: string | null;

  @ApiProperty()
  enabled!: boolean;

  @ApiPropertyOptional()
  lifecycleStatus?: string | null;

  @ApiProperty({ type: () => [String] })
  mode!: string[];

  @ApiPropertyOptional()
  contextWindow?: number | null;

  @ApiPropertyOptional()
  maxOutputTokens?: number | null;

  @ApiProperty()
  openSource!: boolean;

  @ApiProperty()
  free!: boolean;

  @ApiProperty({ type: () => ProviderSummary })
  provider!: ProviderSummary;

  @ApiPropertyOptional({ type: () => ModelSeriesSummary })
  modelSeries?: ModelSeriesSummary | null;
}

class ImportProviderDto {
  @ApiProperty()
  slug!: string;

  @ApiProperty()
  name!: string;

  @ApiPropertyOptional()
  status?: string;

  @ApiPropertyOptional()
  websiteUrl?: string;

  @ApiPropertyOptional()
  docsUrl?: string;

  @ApiPropertyOptional()
  supportsByok?: boolean;
}

class ImportPricingDto {
  @ApiPropertyOptional()
  currency?: string;

  @ApiPropertyOptional()
  inputPricePerMillionTokens?: number;

  @ApiPropertyOptional()
  outputPricePerMillionTokens?: number;

  @ApiPropertyOptional()
  cachedInputPricePerMillionTokens?: number;

  @ApiPropertyOptional()
  imageInputPrice?: number;

  @ApiPropertyOptional()
  imageOutputPrice?: number;

  @ApiPropertyOptional()
  audioInputPrice?: number;

  @ApiPropertyOptional()
  audioOutputPrice?: number;

  @ApiPropertyOptional()
  requestPrice?: number;

  @ApiPropertyOptional()
  region?: string;

  @ApiPropertyOptional()
  sourceUrl?: string;
}

class ImportModelDto {
  @ApiProperty()
  providerSlug!: string;

  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty()
  displayName!: string;

  @ApiPropertyOptional()
  status?: string;

  @ApiPropertyOptional()
  lifecycleStatus?: string;

  @ApiPropertyOptional()
  shortDescription?: string;

  @ApiPropertyOptional()
  contextWindow?: number;

  @ApiPropertyOptional()
  maxOutputTokens?: number;

  @ApiPropertyOptional()
  openSource?: boolean;

  @ApiPropertyOptional()
  free?: boolean;

  @ApiPropertyOptional({ type: () => [String] })
  mode?: string[];

  @ApiPropertyOptional()
  categorySlug?: string;

  @ApiPropertyOptional({ type: () => [String] })
  inputModalities?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  outputModalities?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  supportedParameters?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  capabilities?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  regions?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  sources?: string[];

  @ApiPropertyOptional({
    oneOf: [
      { $ref: '#/components/schemas/ImportPricingDto' },
      {
        type: 'array',
        items: { $ref: '#/components/schemas/ImportPricingDto' },
      },
    ],
  })
  pricing?: ImportPricingDto | ImportPricingDto[];
}

class ImportCatalogRequest {
  @ApiProperty({ type: () => [ImportProviderDto] })
  providers!: ImportProviderDto[];

  @ApiProperty({ type: () => [ImportModelDto] })
  models!: ImportModelDto[];
}

class ImportCatalogResponse {
  @ApiProperty()
  ok!: boolean;

  @ApiPropertyOptional()
  importedProviders?: number;

  @ApiPropertyOptional()
  importedModels?: number;
}

@ApiTags('providers')
@Controller({ path: 'providers', version: '1' })
class OpenAPIProvidersController {
  @Get()
  @ApiOperation({ summary: 'List enabled providers in the discovery catalog.' })
  @ApiOkResponse({ type: [ProviderCatalogEntry] })
  list() {
    return undefined;
  }

  @Get(':slug/models')
  @ApiOperation({
    summary: 'List enabled models for one enabled provider.',
  })
  @ApiOkResponse({ type: ProviderModelsListResponse })
  listModels(@Param('slug') _slug: string) {
    return undefined;
  }

  @Get(':slug')
  @ApiOperation({
    summary: 'Return one enabled provider and its enabled models.',
  })
  @ApiOkResponse({ type: ProviderDetailResponse })
  detail(@Param('slug') _slug: string) {
    return undefined;
  }

  @ApiBearerAuth()
  @Get('admin/list')
  @ApiOperation({
    summary: 'List providers for admin management, including disabled ones.',
  })
  @ApiOkResponse({ type: [ProviderCatalogEntry] })
  listAllForAdmin() {
    return undefined;
  }

  @ApiBearerAuth()
  @Post('admin/upsert')
  @ApiOperation({ summary: 'Create or update a provider definition.' })
  @ApiBody({ type: UpsertProviderDto })
  @ApiOkResponse({ type: ProviderCatalogEntry })
  upsert() {
    return undefined;
  }

  @ApiBearerAuth()
  @Post('admin/enabled')
  @ApiOperation({ summary: 'Enable or disable a provider.' })
  @ApiBody({
    schema: {
      type: 'object',
      required: ['slug', 'enabled'],
      properties: { slug: { type: 'string' }, enabled: { type: 'boolean' } },
    },
  })
  @ApiOkResponse({ type: ProviderCatalogEntry })
  setEnabled() {
    return undefined;
  }

  @ApiBearerAuth()
  @Post('admin/connection-profile/upsert')
  @ApiOperation({
    summary: 'Create or update provider connection profile metadata.',
  })
  @ApiBody({ type: UpsertProviderConnectionProfileDto })
  @ApiOkResponse({ type: ProviderCatalogEntry })
  upsertConnectionProfile() {
    return undefined;
  }

  @ApiBearerAuth()
  @Get('admin/:slug/connection-profile')
  @ApiOperation({ summary: 'Return provider connection profile metadata.' })
  @ApiOkResponse({ type: ProviderCatalogEntry })
  getConnectionProfile(@Param('slug') _slug: string) {
    return undefined;
  }

  @ApiBearerAuth()
  @Post('admin/:slug/logo')
  @ApiConsumes('multipart/form-data')
  @ApiOperation({ summary: 'Upload a provider logo image.' })
  @ApiBody({ type: UploadFileDto })
  @ApiOkResponse({ type: ProviderCatalogEntry })
  uploadProviderLogo(@Param('slug') _slug: string) {
    return undefined;
  }

  @Get('assets/:bucket/:fileName')
  @ApiOperation({ summary: 'Serve a provider or model-series image asset.' })
  @ApiOkResponse({
    description: 'Binary asset response.',
    schema: { type: 'string', format: 'binary' },
  })
  getAsset(
    @Param('bucket') _bucket: string,
    @Param('fileName') _fileName: string,
  ) {
    return undefined;
  }
}

@ApiTags('models')
@Controller({ path: 'models', version: '1' })
class OpenAPICatalogController {
  @Get()
  @ApiOperation({
    summary: 'List catalog models with provider, pricing, and filter metadata.',
  })
  @ApiOkResponse({ type: CatalogListResponse })
  @ApiQuery({ name: 'provider', required: false, type: String })
  @ApiQuery({ name: 'mode', required: false, type: String })
  @ApiQuery({
    name: 'page',
    required: false,
    type: Number,
    schema: { minimum: 1 },
  })
  @ApiQuery({
    name: 'limit',
    required: false,
    type: Number,
    schema: { minimum: 1, maximum: 100 },
  })
  list(@Query() _query: ModelsQueryDto) {
    return undefined;
  }

  @Get('compare')
  @ApiOperation({ summary: 'Compare multiple models by canonical id.' })
  @ApiQuery({
    name: 'ids',
    required: true,
    type: String,
    description: 'Comma-separated canonical model ids.',
  })
  @ApiOkResponse({ type: [CatalogModelDetailResponse] })
  compare(@Query('ids') _ids: string) {
    return undefined;
  }

  @Get('recent')
  @ApiOperation({ summary: 'Return recently released enabled models.' })
  @ApiOkResponse({ type: [CatalogModelListItem] })
  recent() {
    return undefined;
  }

  @Get('free')
  @ApiOperation({ summary: 'Return enabled free-to-use models.' })
  @ApiOkResponse({ type: [CatalogModelDetailResponse] })
  free() {
    return undefined;
  }

  @Get('deprecated')
  @ApiOperation({ summary: 'Return deprecated enabled models.' })
  @ApiOkResponse({ type: [CatalogModelDetailResponse] })
  deprecated() {
    return undefined;
  }

  @Get(':canonicalId')
  @ApiOperation({ summary: 'Return one enabled model by canonical id.' })
  @ApiOkResponse({ type: CatalogModelDetailResponse })
  getOne(@Param('canonicalId') _canonicalId: string) {
    return undefined;
  }
}

@ApiTags('admin')
@ApiBearerAuth()
@Controller({ path: 'admin/models', version: '1' })
class OpenAPICatalogAdminController {
  @Post('upsert')
  @ApiOperation({
    summary: 'Create or update a model row and optional pricing snapshot.',
  })
  @ApiBody({ type: UpsertModelDto })
  @ApiOkResponse({ type: AdminModelRecord })
  upsert() {
    return undefined;
  }

  @Get('audit-log')
  @ApiOperation({ summary: 'List recent admin audit log entries.' })
  @ApiOkResponse({ type: [AuditLogEntry] })
  audit() {
    return undefined;
  }

  @Get()
  @ApiOperation({ summary: 'List models for admin management.' })
  @ApiOkResponse({ type: [AdminModelRecord] })
  listAll() {
    return undefined;
  }

  @Post(':canonicalId/enabled')
  @ApiOperation({ summary: 'Enable or disable one model by canonical id.' })
  @ApiBody({ type: BooleanToggleDto })
  @ApiOkResponse({ type: AdminModelRecord })
  setModelEnabled(@Param('canonicalId') _canonicalId: string) {
    return undefined;
  }

  @Get('series')
  @ApiOperation({ summary: 'List model series rows.' })
  @ApiOkResponse({ type: [ModelSeriesRecord] })
  listSeries() {
    return undefined;
  }

  @Post('series/upsert')
  @ApiOperation({ summary: 'Create or update a model series.' })
  @ApiBody({ type: UpsertModelSeriesDto })
  @ApiOkResponse({ type: ModelSeriesRecord })
  upsertSeries() {
    return undefined;
  }

  @Post('series/:slug/logo')
  @ApiConsumes('multipart/form-data')
  @ApiOperation({ summary: 'Upload a model-series logo image.' })
  @ApiBody({ type: UploadFileDto })
  @ApiOkResponse({ type: ModelSeriesRecord })
  uploadSeriesLogo(@Param('slug') _slug: string) {
    return undefined;
  }
}

@ApiTags('settings')
@Controller({ path: 'settings/display', version: '1' })
class OpenAPIDisplaySettingsController {
  @Get()
  @ApiOperation({ summary: 'Return provider-catalog display settings.' })
  @ApiOkResponse({ type: DisplaySettingsResponse })
  getDisplaySettings() {
    return undefined;
  }

  @ApiBearerAuth()
  @Put()
  @ApiOperation({ summary: 'Update provider-catalog display settings.' })
  @ApiBody({ type: UpdateDisplaySettingsDto })
  @ApiOkResponse({ type: DisplaySettingsResponse })
  updateDisplaySettings() {
    return undefined;
  }
}

@ApiTags('facets')
@Controller({ version: '1' })
class OpenAPIFacetsController {
  @Get('facets')
  @ApiOperation({
    summary:
      'Return the aggregated filter option payload used by the catalog UI.',
  })
  @ApiOkResponse({ type: FacetsResponse })
  all() {
    return undefined;
  }

  @Get('categories')
  @ApiOkResponse({ type: [FacetOption] })
  categories() {
    return undefined;
  }

  @Get('capabilities')
  @ApiOkResponse({ type: [FacetOption] })
  capabilities() {
    return undefined;
  }

  @Get('supported-parameters')
  @ApiOkResponse({ type: [FacetOption] })
  supportedParameters() {
    return undefined;
  }

  @Get('modalities')
  @ApiOkResponse({ type: [FacetOption] })
  modalities() {
    return undefined;
  }

  @Get('regions')
  @ApiOkResponse({ type: [FacetOption] })
  regions() {
    return undefined;
  }
}

@ApiTags('indexing')
@ApiBearerAuth()
@Controller({ path: 'admin/indexing', version: '1' })
class OpenAPIIndexingController {
  @Post('rebuild')
  @ApiOperation({ summary: 'Queue a full catalog reindex job.' })
  @ApiOkResponse({ type: SimpleAckResponse })
  rebuild() {
    return undefined;
  }

  @Post('model')
  @ApiOperation({ summary: 'Queue one model for reindexing.' })
  @ApiBody({ type: IndexOneDto })
  @ApiOkResponse({ type: SimpleAckResponse })
  one() {
    return undefined;
  }
}

@ApiTags('import')
@ApiBearerAuth()
@Controller({ path: 'admin/import', version: '1' })
class OpenAPIImportController {
  @Post('catalog-json')
  @ApiOperation({ summary: 'Import providers and models from JSON payload.' })
  @ApiBody({ type: ImportCatalogRequest })
  @ApiOkResponse({ type: ImportCatalogResponse })
  importCatalog() {
    return undefined;
  }
}

@Module({
  controllers: [
    OpenAPIProvidersController,
    OpenAPICatalogController,
    OpenAPICatalogAdminController,
    OpenAPIDisplaySettingsController,
    OpenAPIFacetsController,
    OpenAPIIndexingController,
    OpenAPIImportController,
  ],
})
export class OpenAPIAppModule {}
