import type { Identity } from "../api/types";
import { CONTACT_LABEL, contactUrl } from "../utils/contacts";

export function ContactLinks({ identities }: { identities: Identity[] }) {
  const links = identities
    .map((identity) => ({ identity, url: contactUrl(identity) }))
    .filter((x): x is { identity: Identity; url: string } => x.url !== null);

  if (links.length === 0) return null;

  return (
    <span className="contact-links">
      {links.map(({ identity, url }) => (
        <a
          key={identity.provider}
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="contact-link"
        >
          {CONTACT_LABEL[identity.provider] ?? identity.provider}
        </a>
      ))}
    </span>
  );
}
