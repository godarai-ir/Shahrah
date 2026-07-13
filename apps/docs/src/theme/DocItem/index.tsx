import type {ReactNode} from 'react';
import OriginalDocItem from '@theme-original/DocItem';

export default function DocItem(props: {children: ReactNode}): ReactNode {
  return <OriginalDocItem {...props} />;
}
