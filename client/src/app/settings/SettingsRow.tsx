export const SettingRow = ({
  title,
  description,
  action
}: {
  title: string;
  description: string;
  action: React.ReactNode;
}) => (
  <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
    <div>
      <h3 className="font-medium">{title}</h3>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
    <div className="w-full md:w-auto">{action}</div>
  </div>
);
