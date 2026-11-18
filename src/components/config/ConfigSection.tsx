import React, { ReactNode } from 'react';
import { css } from '@emotion/css';
import { useTheme2 } from '@grafana/ui';

type Props = {
  id: string;
  title: string;
  description?: ReactNode;
  children: ReactNode;
};

// A titled config section (h3 header + optional description + content), matching
// the modern Grafana data source config layout (cf. ClickHouse config-v2).
export const ConfigSection = ({ id, title, description, children }: Props) => {
  const { colors, typography, spacing } = useTheme2();
  const styles = {
    title: css({ margin: 0 }),
    description: css({
      marginTop: spacing(0.5),
      marginBottom: 0,
      ...typography.sm,
      color: colors.text.secondary,
    }),
    content: css({ marginTop: spacing(2) }),
    section: css({ marginBottom: spacing(3) }),
  };
  return (
    <div id={id} className={styles.section}>
      <h3 className={styles.title}>{title}</h3>
      {description && <p className={styles.description}>{description}</p>}
      <div className={styles.content}>{children}</div>
    </div>
  );
};
