import { useSection } from "../auth/Sections";
import { t } from "../i18n";

export function VenueSelect({
  value,
  onChange,
  savedVenue,
}: {
  value: string;
  onChange: (venue: string) => void;
  savedVenue?: string;
}) {
  const { current } = useSection();
  const venues = current.venues ?? [];
  return (
    <label className="form__field">
      <span>{t("Зал")}</span>
      <select className="venue-select" value={value} onChange={(event) => onChange(event.target.value)}>
        <option value="">{t("Место не указано")}</option>
        {savedVenue && !venues.includes(savedVenue) && <option value={savedVenue}>{savedVenue}</option>}
        {venues.map((venue) => (
          <option value={venue} key={venue}>
            {venue}
          </option>
        ))}
      </select>
    </label>
  );
}
