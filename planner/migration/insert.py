import os
import re
import zipfile
import xml.etree.ElementTree as ET
from datetime import datetime, timedelta

def parse_excel_sheet1(file_path):
    with zipfile.ZipFile(file_path) as z:
        shared_strings = []
        if 'xl/sharedStrings.xml' in z.namelist():
            tree = ET.fromstring(z.read('xl/sharedStrings.xml'))
            for si in tree.findall('.//{http://schemas.openxmlformats.org/spreadsheetml/2006/main}si'):
                text_parts = [elem.text for elem in si.iter() if elem.tag.endswith('}t') and elem.text]
                shared_strings.append(''.join(text_parts))

        tree = ET.fromstring(z.read('xl/worksheets/sheet1.xml'))
        rows_data = []
        for row in tree.findall('.//{http://schemas.openxmlformats.org/spreadsheetml/2006/main}row'):
            row_num = int(row.attrib.get('r'))
            row_dict = {'__row__': row_num}
            for c in row.findall('{http://schemas.openxmlformats.org/spreadsheetml/2006/main}c'):
                r = c.attrib.get('r')
                t = c.attrib.get('t')
                v = c.find('{http://schemas.openxmlformats.org/spreadsheetml/2006/main}v')
                val = v.text if v is not None else None
                if t == 's' and val is not None:
                    val = shared_strings[int(val)]
                elif t == 'str':
                    val = val if val is not None else ""
                col = re.match(r'([A-Z]+)', r).group(1)
                row_dict[col] = val
            rows_data.append(row_dict)
        return rows_data

def col_val(r, col):
    v = r.get(col)
    if v is None or v == "" or str(v).strip() == "":
        return None
    return str(v).strip()

def parse_date_val(s):
    if s is None or s == "" or str(s).strip() == "":
        return "null"
    s = str(s).strip()
    if s.lower() == 'null':
        return "null"
    
    # Check if float serial number from Excel
    try:
        f = float(s)
        d = datetime(1899, 12, 30) + timedelta(days=f)
        return f"'{d.strftime('%Y-%m-%d')}'"
    except ValueError:
        pass

    # Check if DD/MM/YYYY or DD/MM/YY format
    m = re.match(r'^(\d{1,2})/(\d{1,2})/(\d{2,4})$', s)
    if m:
        day, month, year = int(m.group(1)), int(m.group(2)), int(m.group(3))
        if year == 226 or year == 26:
            year = 2026
        return f"'{year:04d}-{month:02d}-{day:02d}'"

    return f"'{s}'"

def format_sql_val(val, val_type):
    if val is None or val == "" or str(val).strip() == "":
        return "null"
    s = str(val).strip()
    if s.lower() == 'null':
        return "null"
    if s.lower() == 'now()':
        return "now()"

    if val_type == 'int':
        return str(int(float(s)))
    elif val_type == 'numeric':
        return f"{float(s):.2f}"
    elif val_type == 'date':
        return parse_date_val(s)
    elif val_type == 'timestamp':
        return f"'{s}'"
    else:
        return f"'{s}'"

def generate_inserts(excel_path, output_sql_path):
    rows_data = parse_excel_sheet1(excel_path)

    plans = []
    plan_items = []
    plan_agendas = []
    plan_notebooks = []
    plan_packages = []

    for r in rows_data:
        row_num = r['__row__']
        if row_num <= 2:  # Skip table names and column headers
            continue

        # plan: A..I
        p_id = col_val(r, 'A')
        if p_id is not None:
            plans.append((
                format_sql_val(p_id, 'int'),
                format_sql_val(col_val(r, 'B'), 'int'),
                format_sql_val(col_val(r, 'C'), 'date'),
                format_sql_val(col_val(r, 'D'), 'date'),
                format_sql_val(col_val(r, 'E'), 'int'),
                format_sql_val(col_val(r, 'F'), 'numeric'),
                format_sql_val(col_val(r, 'G'), 'timestamp'),
                format_sql_val(col_val(r, 'H'), 'timestamp'),
                format_sql_val(col_val(r, 'I'), 'timestamp')
            ))

        # plan_item: J..Q
        pi_id = col_val(r, 'J')
        if pi_id is not None:
            plan_items.append((
                format_sql_val(pi_id, 'int'),
                format_sql_val(col_val(r, 'K'), 'int'),
                format_sql_val(col_val(r, 'L'), 'int'),
                format_sql_val(col_val(r, 'M'), 'int'),
                format_sql_val(col_val(r, 'N'), 'numeric'),
                format_sql_val(col_val(r, 'O'), 'timestamp'),
                format_sql_val(col_val(r, 'P'), 'timestamp'),
                format_sql_val(col_val(r, 'Q'), 'timestamp')
            ))

        # plan_agenda: R..W (plan_type = 1)
        pa_id = col_val(r, 'R')
        if pa_id is not None:
            plan_agendas.append((
                format_sql_val(pa_id, 'int'),
                format_sql_val(col_val(r, 'S'), 'int'),
                format_sql_val(col_val(r, 'T'), 'int'),
                format_sql_val(col_val(r, 'U'), 'int'),
                format_sql_val(col_val(r, 'V'), 'int'),
                format_sql_val(col_val(r, 'W'), 'int')
            ))

        # plan_notebook: X..Z (plan_type = 3)
        pn_id = col_val(r, 'X')
        if pn_id is not None:
            plan_notebooks.append((
                format_sql_val(pn_id, 'int'),
                format_sql_val(col_val(r, 'Y'), 'int'),
                format_sql_val(col_val(r, 'Z'), 'int')
            ))

        # plan_package: AA..AC (plan_type = 2)
        pp_id = col_val(r, 'AA')
        if pp_id is not None:
            plan_packages.append((
                format_sql_val(pp_id, 'int'),
                format_sql_val(col_val(r, 'AB'), 'int'),
                format_sql_val(col_val(r, 'AC'), 'date')
            ))

    sql_lines = [
        "-- Active: 1790015123887@@127.0.0.1@5433@planner@planner",
        "create schema if not exists planner;",
        "",
        "set search_path to planner;",
        "",
        "-- ============================================================================",
        "-- 1. Table: plan",
        "-- ============================================================================",
        "insert into plan (id, customer_id, plan_start, plan_end, plan_type, price, created_at, updated_at, deleted_at) values"
    ]

    plan_val_strs = [f"({', '.join(p)})" for p in plans]
    sql_lines.append(",\n".join(plan_val_strs) + ";")
    sql_lines.append("")

    sql_lines.extend([
        "-- ============================================================================",
        "-- 2. Table: plan_item",
        "-- ============================================================================",
        "insert into plan_item (id, plan_id, service_id, order_index, price, created_at, updated_at, deleted_at) values"
    ])
    item_val_strs = [f"({', '.join(pi)})" for pi in plan_items]
    sql_lines.append(",\n".join(item_val_strs) + ";")
    sql_lines.append("")

    sql_lines.extend([
        "-- ============================================================================",
        "-- 3. Table: plan_agenda (specialization: plan_type = 1)",
        "-- ============================================================================",
        "insert into plan_agenda (id, recurrence, week_day, term_type, payment_day, monthly_service_limit) values"
    ])
    agenda_val_strs = [f"({', '.join(pa)})" for pa in plan_agendas]
    sql_lines.append(",\n".join(agenda_val_strs) + ";")
    sql_lines.append("")

    sql_lines.extend([
        "-- ============================================================================",
        "-- 4. Table: plan_notebook (specialization: plan_type = 3)",
        "-- ============================================================================",
        "insert into plan_notebook (id, payment_day, payment_term) values"
    ])
    notebook_val_strs = [f"({', '.join(pn)})" for pn in plan_notebooks]
    sql_lines.append(",\n".join(notebook_val_strs) + ";")
    sql_lines.append("")

    sql_lines.extend([
        "-- ============================================================================",
        "-- 5. Table: plan_package (specialization: plan_type = 2)",
        "-- ============================================================================",
        "insert into plan_package (id, quantity, payment_date) values"
    ])
    package_val_strs = [f"({', '.join(pp)})" for pp in plan_packages]
    sql_lines.append(",\n".join(package_val_strs) + ";")
    sql_lines.append("")

    sql_lines.extend([
        "-- ============================================================================",
        "-- 6. Sync sequences",
        "-- ============================================================================",
        "select setval('planner.plan_id_seq', coalesce((select max(id) from planner.plan), 1));",
        "select setval('planner.plan_item_id_seq', coalesce((select max(id) from planner.plan_item), 1));",
        ""
    ])

    sql_content = "\n".join(sql_lines)
    with open(output_sql_path, 'w', encoding='utf-8') as f:
        f.write(sql_content)

    print(f"Generated {output_sql_path} successfully ({len(plans)} plans, {len(plan_items)} items, {len(plan_agendas)} agendas, {len(plan_notebooks)} notebooks, {len(plan_packages)} packages).")

if __name__ == '__main__':
    base_dir = os.path.dirname(os.path.abspath(__file__))
    excel_file = os.path.join(base_dir, 'plan.xlsx')
    output_file = os.path.join(base_dir, 'insert.sql')
    generate_inserts(excel_file, output_file)
