ROBOCOPY "D:\Prowler\projects\CNC-list-creator\del\orig" "D:\Prowler\projects\CNC-list-creator\del\for_test" /E /PURGE
ROBOCOPY "D:\Prowler\projects\CNC-list-creator\proj\bin" "D:\Prowler\projects\CNC-list-creator\del\for_test" "ListMaker.exe" "listMaker_settings.xml"
