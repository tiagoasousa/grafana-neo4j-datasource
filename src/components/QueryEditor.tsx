import React, { ChangeEvent, PureComponent } from 'react';
import { CodeEditor, InlineField, InlineFieldRow, InlineFormLabel, Input, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { Format, MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const formats = [
  { label: 'Table', value: Format.Table, description: 'Table view' },
  { label: 'Node Graph', value: Format.NodeGraph, description: 'Node Graph view' },
] as Array<SelectableValue<Format>>;

export class QueryEditor extends PureComponent<Props> {
  onCypherQueryChange = (value: string | undefined) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, cypherQuery: value || '' });
    onRunQuery();
  };

  onFormatChanged = (selected: SelectableValue<Format>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, format: selected.value || Format.Table });
    onRunQuery();
  };

  onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, database: event.currentTarget.value });
  };

  render() {
    const { query } = this.props;
    return (
      <div>
        <CodeEditor
          height="240px"
          language="cypher"
          value={query.cypherQuery || ''}
          onEditorDidMount={(editor) => {
            editor.onDidChangeModelContent(() => this.onCypherQueryChange(editor.getValue()));
          }}
          monacoOptions={{ minimap: { enabled: false }, automaticLayout: true }}
        />
        <InlineFieldRow>
          <InlineField label="Database" labelWidth={7} tooltip="Overrides the data source database for this query only">
            <Input
              className="width-14"
              value={query.database || ''}
              placeholder="defaults to the data source setting"
              onChange={this.onDatabaseChange}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineFormLabel width={7}>Format</InlineFormLabel>
          <Select
            className="width-14"
            value={formats.find((f) => f.value === query.format) || formats[0]}
            options={formats}
            onChange={this.onFormatChanged}
          />
        </InlineFieldRow>
      </div>
    );
  }
}
