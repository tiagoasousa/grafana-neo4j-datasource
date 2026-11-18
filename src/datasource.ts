import {
  CoreApp,
  DataFrame,
  DataFrameView,
  DataQueryRequest,
  DataSourceInstanceSettings,
  MetricFindValue,
  ScopedVars,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { firstValueFrom } from 'rxjs';
import { DEFAULT_QUERY, MyDataSourceOptions, MyQuery } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    return {
      ...query,
      cypherQuery: getTemplateSrv().replace(query.cypherQuery || '', scopedVars),
    };
  }

  filterQuery(query: MyQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.cypherQuery;
  }

  // Used for dashboard template variables (VariableQueryEditor can reuse QueryEditor).
  async metricFindQuery(query: MyQuery, options: any): Promise<MetricFindValue[]> {
    const evaluated = this.applyTemplateVariables(query, options.scopedVars);
    const request = {
      targets: [{ ...evaluated, refId: 'metricFindQuery' }],
      range: options.range,
      rangeRaw: options.rangeRaw,
    } as DataQueryRequest<MyQuery>;

    const res = await firstValueFrom(this.query(request));
    if (!res || !res.data || res.data.length === 0) {
      return [];
    }
    const frame = res.data[0] as DataFrame;
    if (!frame || frame.fields.length === 0) {
      return [];
    }
    const field = frame.fields[0].name;
    return new DataFrameView(frame).map((item) => ({ text: item[field] }));
  }
}
