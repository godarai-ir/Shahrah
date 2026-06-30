import type {ReactNode} from 'react';
import OriginalDocItem from '@theme-original/DocItem';

import FeedbackBlock from '@site/src/theme/DocItem/FeedbackBlock';

export default function DocItem(props: {children: ReactNode}): ReactNode {
  return (
    <>
      <OriginalDocItem {...props} />
      <FeedbackBlock />
    </>
  );
}
