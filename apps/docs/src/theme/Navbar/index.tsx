import type {ReactNode} from 'react';
import OriginalNavbar from '@theme-original/Navbar';
import Link from '@docusaurus/Link';

export default function NavbarWrapper(props: Record<string, unknown>): ReactNode {
  return (
    <div className="navbar-with-inline-search">
      <OriginalNavbar {...props} />
      <Link className="navbar-inline-search" to="/search">
        جستجو
      </Link>
    </div>
  );
}
