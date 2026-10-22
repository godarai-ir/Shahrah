import { Pebbles } from './FordArt';
import { MediaSlot } from './MediaSlot';
import { FeatureRow } from './patterns';

export function Collab() {
  return (
    <FeatureRow
      id="collab"
      title="کار میان همکارها و دستیار دست‌به‌دست می‌شود"
      copy="دستیار پیش‌نویس را آماده می‌کند، همکارتان نگاهی می‌اندازد و کار از همان‌جا که تیم‌تان هست پیش می‌رود: بله، ایتا، واتساپ، تلگرام، Jira یا Notion."
      visual={
        <div className="relative w-full max-w-lg">
          <div className="overflow-hidden rounded-3xl">
            <MediaSlot id="collabPhoto" className="h-full rounded-3xl" />
          </div>
          <Pebbles className="mt-6 w-48 sm:w-56" />
        </div>
      }
    />
  );
}
