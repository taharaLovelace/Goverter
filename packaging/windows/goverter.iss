#define MyAppName "Goverter"
#define MyAppPublisher "Hugo Tahara Menegatti"
#define MyAppURL "https://github.com/taharaLovelace/Goverter"
#ifndef MyAppVersion
  #define MyAppVersion "dev"
#endif
#ifndef SourceDir
  #define SourceDir "..\..\dist"
#endif

[Setup]
AppId={{6C8EA70A-724D-4919-B14E-442354EA188C}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}/issues
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={localappdata}\Programs\Goverter
DefaultGroupName=Goverter
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
MinVersion=10.0
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir={#SourceDir}\installer
OutputBaseFilename=Goverter-{#MyAppVersion}-windows-x64
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes
LicenseFile=..\..\LICENSE
UninstallDisplayIcon={app}\goverter.exe

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "addtopath"; Description: "Add Goverter to the current user's PATH"; Flags: checkedonce

[Files]
Source: "{#SourceDir}\goverter.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\tools\ffmpeg.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "{#SourceDir}\tools\ffprobe.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "{#SourceDir}\third_party\ffmpeg\LICENSE"; DestDir: "{app}\licenses\ffmpeg"; Flags: ignoreversion
Source: "{#SourceDir}\third_party\ffmpeg\README.txt"; DestDir: "{app}\licenses\ffmpeg"; Flags: ignoreversion
Source: "{#SourceDir}\third_party\ffmpeg\tools.lock.json"; DestDir: "{app}\licenses\ffmpeg"; Flags: ignoreversion
Source: "{#SourceDir}\third_party\go\cobra\LICENSE"; DestDir: "{app}\licenses\cobra"; Flags: ignoreversion
Source: "{#SourceDir}\third_party\go\pflag\LICENSE"; DestDir: "{app}\licenses\pflag"; Flags: ignoreversion
Source: "{#SourceDir}\third_party\go\mousetrap\LICENSE"; DestDir: "{app}\licenses\mousetrap"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}\licenses\goverter"; Flags: ignoreversion
Source: "..\..\THIRD_PARTY_NOTICES.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\CHANGELOG.md"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; \
    ValueData: "{code:PathWithGoverter}"; Flags: preservestringtype; Tasks: addtopath

[Code]
function PathWithGoverter(Param: String): String;
var
  CurrentPath: String;
  AppPath: String;
  PaddedPath: String;
begin
  AppPath := ExpandConstant('{app}');
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', CurrentPath) then
    CurrentPath := '';

  PaddedPath := ';' + CurrentPath + ';';
  if Pos(';' + Lowercase(AppPath) + ';', Lowercase(PaddedPath)) > 0 then
    Result := CurrentPath
  else if CurrentPath = '' then
    Result := AppPath
  else
    Result := CurrentPath + ';' + AppPath;
end;

procedure RemoveGoverterFromPath;
var
  CurrentPath: String;
  AppPath: String;
  PaddedPath: String;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', CurrentPath) then
    exit;

  AppPath := ExpandConstant('{app}');
  PaddedPath := ';' + CurrentPath + ';';
  StringChangeEx(PaddedPath, ';' + AppPath + ';', ';', True);

  while (Length(PaddedPath) > 0) and (PaddedPath[1] = ';') do
    Delete(PaddedPath, 1, 1);
  while (Length(PaddedPath) > 0) and (PaddedPath[Length(PaddedPath)] = ';') do
    Delete(PaddedPath, Length(PaddedPath), 1);

  RegWriteExpandStringValue(HKCU, 'Environment', 'Path', PaddedPath);
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
    RemoveGoverterFromPath;
end;
