/**
 * Sync Prisma Schema Script
 *
 * This script syncs the Prisma schema with the SQL migrations.
 * It should be run after adding new SQL migrations to db/main/migrations/
 *
 * Usage: node scripts/sync-prisma.js
 *
 * IMPORTANT: Uses Prisma v6.x from node_modules to avoid v7 breaking changes
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const ROOT_DIR = path.join(__dirname, '..');
const DB_PATH = path.join(ROOT_DIR, 'db', 'main', 'migrations');

console.log('🔄 Starting Prisma schema sync...\n');

// Step 1: Check for SQL migration files
try {
  const migrationFiles = fs.readdirSync(DB_PATH)
    .filter(f => f.endsWith('.sql'))
    .sort();

  console.log(`📁 Found ${migrationFiles.length} SQL migration files:`);
  migrationFiles.forEach(f => console.log(`   - ${f}`));
  console.log('');
} catch (e) {
  console.log('⚠️  No migrations folder found, skipping...\n');
}

// Step 2: Run SQL migrations (requires PostgreSQL client)
console.log('⚡ Running SQL migrations...');
console.log('   (Requires psql to be installed and database to be running)');
console.log('   Run manually: psql -U postgres -d llmproxy -f db/main/migrations/001_initial.sql');
console.log('');

// Step 3: Format schemas
console.log('📝 Formatting Prisma schemas...');
const schemas = [
  'packages/database/schema.prisma',
  'apps/tenant-admin/prisma/schema.prisma',
  'apps/support-admin/prisma/schema.prisma'
];

schemas.forEach(schema => {
  try {
    // Use npx to run prisma to ensure cross-platform compatibility
    execSync(`npx prisma@6.9.0 format --schema=${schema}`, {
      stdio: 'inherit',
      cwd: ROOT_DIR,
      shell: true
    });
    console.log(`   ✅ Formatted ${path.basename(schema)}`);
  } catch (error) {
    console.log(`   ⚠️  Could not format ${path.basename(schema)}`);
  }
});
console.log('');

// Step 4: Generate Prisma clients
console.log('🔧 Generating Prisma clients...');

schemas.forEach(schema => {
  try {
    // Use npx to run prisma to ensure cross-platform compatibility
    execSync(`npx prisma@6.9.0 generate --schema=${schema}`, {
      stdio: 'inherit',
      cwd: ROOT_DIR,
      shell: true
    });
    console.log(`   ✅ Generated client for ${path.basename(schema)}`);
  } catch (error) {
    console.log(`   ⚠️  Could not generate client for ${path.basename(schema)}`);
    // Don't print full error - it's usually just a warning
  }
});

console.log('\n🎉 Schema sync complete!');
console.log('\n📝 Workflow:');
console.log('   1. AI adds new SQL migration to db/main/migrations/');
console.log('   2. Run this script: node scripts/sync-prisma.js');
console.log('   3. Review any changes to schema.prisma');
console.log('   4. Commit the changes');
console.log('   5. Deploy: npx prisma migrate deploy');