import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();

  return (
    <Layout
      title={siteConfig.title}
      description="Shahrah platform documentation">
      <main style={{padding: '6rem 1.5rem'}}>
        <div style={{margin: '0 auto', maxWidth: 960}}>
          <p
            style={{
              color: '#0f766e',
              fontWeight: 700,
              letterSpacing: '0.2em',
              textTransform: 'uppercase',
            }}>
            Shahrah
          </p>
          <h1
            style={{
              fontSize: 'clamp(3rem, 7vw, 5rem)',
              lineHeight: 0.95,
              margin: '1rem 0',
            }}>
            Documentation
          </h1>
          <p style={{fontSize: '1.125rem', lineHeight: 1.8, maxWidth: 720}}>
            Docs for the Shahrah control plane, tenant model, billing flows,
            SDKs, and platform architecture.
          </p>
          <div
            style={{
              display: 'flex',
              gap: '1rem',
              marginTop: '2rem',
              flexWrap: 'wrap',
            }}>
            <Link className="button button--primary button--lg" to="/intro">
              Read the intro
            </Link>
            <Link
              className="button button--secondary button--lg"
              href="https://sharah.ir/blog">
              Visit the blog
            </Link>
          </div>
        </div>
      </main>
    </Layout>
  );
}
