import React from 'react';
import { css } from '@emotion/css';
import { Box, Icon, LinkButton, Space, Text, useStyles2 } from '@grafana/ui';

export type SidebarHeader = { label: string; id: string; isOptional?: boolean };

// Sticky left navigation listing the config sections; clicking smooth-scrolls to
// the matching section (cf. ClickHouse config-v2 LeftSidebar).
export const LeftSidebar = ({ headers }: { headers: SidebarHeader[] }) => {
  const styles = useStyles2(getStyles);
  const scrollTo = (id: string) => (e: React.MouseEvent) => {
    e.preventDefault();
    const target = document.getElementById(id);
    if (target) {
      const top = target.getBoundingClientRect().top + window.scrollY - 60;
      window.scrollTo({ top, behavior: 'smooth' });
    }
  };
  return (
    <div className={styles.sidebar}>
      <Text element="h4">Connect data source</Text>
      <Box paddingTop={2}>
        {headers.map((header) => (
          <div key={header.id} data-testid={`${header.label}-sidebar`}>
            <Icon name="circle" size="xs" />
            <LinkButton variant="secondary" fill="text" onClick={scrollTo(header.id)}>
              <div className={styles.sidebarText}>
                <div className={styles.sidebarLabel}>{header.label}</div>
                {header.isOptional && (
                  <Text color="secondary" variant="sm">
                    optional
                  </Text>
                )}
              </div>
            </LinkButton>
            <Space v={1} />
          </div>
        ))}
      </Box>
    </div>
  );
};

const getStyles = () => ({
  sidebar: css({
    position: 'sticky',
    top: 0,
    alignSelf: 'flex-start',
    minWidth: '200px',
  }),
  sidebarText: css({ display: 'flex', flexDirection: 'column' }),
  sidebarLabel: css({ display: 'flex', alignItems: 'center', lineHeight: 1 }),
});
