import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError, map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class SourceService {

  constructor(private http: HttpClient,
    private logging: LoggingService) {
    this.logger = logging.get('SourceService');
    this.current = null;
  }

  current: SourceInterface;
  private logger: Logger;

  list(): Observable<SourceInterface[]> {
    const url = `${environment.apiURL}sources`;
    this.logger.log(`Retrieving sources`, url);
    return this.http.get<sourceList>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved sources', res)),
        map(res => res.items.map(item => new SourceInterface(this.http, this.logging, item.name, item.status))),
        catchError(this.logger.handleError<SourceInterface[]>(`list()`))
      );
  }
}

export class SourceInterface {
  constructor(private http: HttpClient,
    logging: LoggingService,
    public name: string,
    public status: string) {
    this.logger = logging.get('SourceInterface');
    this.uriName = encodeURIComponent(this.name);
  }

  private uriName: string;
  private logger: Logger;

  get(): Observable<SourceDetails> {
    const url = `${environment.apiURL}sources/${this.uriName}`;
    this.logger.log(`Retrieving source details for ${this.name}`, url);
    return this.http.get<SourceDetails>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved source details', res)),
        catchError(this.logger.handleError<SourceDetails>(`'${this.name}'->get()`))
      );
  }
}

export interface SourceDetails {
}

interface sourceListItem {
  name: string;
  status: string;
}

interface sourceList {
  items: sourceListItem[];
}